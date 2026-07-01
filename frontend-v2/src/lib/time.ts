// Time and timezone helpers for events.
//
// An upcoming event is stored as a local calendar date, local wall-clock times,
// and an IANA timezone name (e.g. "America/Los_Angeles") — deliberately not a
// single UTC instant. The intent is "7:00 PM in Los Angeles on this date", and
// anchoring to wall-clock + zone keeps that fixed for every viewer and survives
// revisions to a zone's UTC offset (DST rules change more often than people
// expect). A UTC instant precomputed at creation time would silently drift if
// those rules later changed, and would lose the originally intended local time.
// Display always formats back in the event's own timezone with a label.

import type { EventListItem } from './api'

/**
 * A small, ordered curated list of common zones offered for manual timezone
 * selection. The full ~400-zone IANA set is overwhelming and unnecessary —
 * these cover where the org operates. `getCommonTimezones` prepends the viewer's
 * own zone when it isn't already listed, so an unlisted zone stays selectable.
 */
const COMMON_TIMEZONES = [
  'America/Los_Angeles',
  'America/Denver',
  'America/Chicago',
  'America/New_York',
  'America/Sao_Paulo',
  'Europe/London',
  'Europe/Paris',
  'Europe/Berlin',
  'Asia/Dubai',
  'Asia/Kolkata',
  'Asia/Singapore',
  'Asia/Tokyo',
  'Australia/Sydney',
  'UTC',
]

/** The creator's browser timezone, used as the default on create. */
export function getBrowserTimezone(): string {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC'
  } catch {
    return 'UTC'
  }
}

/**
 * The curated list for manual timezone selection, always including the viewer's
 * browser zone so it is selectable even when it isn't one of the common zones.
 */
export function getCommonTimezones(): string[] {
  const browser = getBrowserTimezone()
  return COMMON_TIMEZONES.includes(browser)
    ? COMMON_TIMEZONES
    : [browser, ...COMMON_TIMEZONES]
}

/**
 * Today's calendar date (YYYY-MM-DD) in the given timezone. Used to compute
 * "today's events" in a defined zone rather than the DB/browser raw date.
 * `now` is injectable so callers can recompute on a timer (e.g. roll over at
 * local midnight). Falls back to the local-zone date if `timezone` is invalid,
 * so a bad stored zone can't throw during render.
 */
export function todayInTimezone(
  timezone: string,
  now: Date = new Date(),
): string {
  const opts: Intl.DateTimeFormatOptions = {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  }
  try {
    // en-CA formats as YYYY-MM-DD.
    return new Intl.DateTimeFormat('en-CA', {
      timeZone: timezone,
      ...opts,
    }).format(now)
  } catch {
    return new Intl.DateTimeFormat('en-CA', opts).format(now)
  }
}

/**
 * Short timezone abbreviation (e.g. "PDT") for the given date in the given zone.
 * We pick noon UTC on the event's date so the abbreviation lands on the correct
 * side of any DST transition for essentially all real-world cases.
 */
export function getZoneAbbreviation(dateStr: string, timezone: string): string {
  if (!timezone) return ''
  const instant = dateStr ? new Date(`${dateStr}T12:00:00Z`) : new Date()
  try {
    const parts = new Intl.DateTimeFormat('en-US', {
      timeZone: timezone,
      timeZoneName: 'short',
    }).formatToParts(instant)
    return parts.find((p) => p.type === 'timeZoneName')?.value ?? timezone
  } catch {
    return timezone
  }
}

/**
 * Parses a "HH:MM" (or "HH:MM:SS") wall-clock string into numeric hours and
 * minutes; null if it doesn't match. The single place the accepted time format
 * is defined — both the 12h label and the minutes-since-midnight math use it.
 */
function parseWallClock(
  time: string,
): { hours: number; minutes: number } | null {
  const match = /^(\d{1,2}):(\d{2})/.exec(time.trim())
  if (!match) return null
  return { hours: Number(match[1]), minutes: Number(match[2]) }
}

/**
 * Formats a 24h "HH:MM" wall-clock string as a 12h label (e.g. "3:00 PM").
 * Returns '' for empty/invalid input.
 */
export function formatWallClock(time: string): string {
  const parsed = parseWallClock(time)
  if (!parsed) return ''
  const { hours, minutes } = parsed
  const period = hours >= 12 ? 'PM' : 'AM'
  const hour12 = hours % 12 === 0 ? 12 : hours % 12
  return `${hour12}:${String(minutes).padStart(2, '0')} ${period}`
}

// The subset of an event's fields needed to render its time range. Taken as a
// single object (rather than four positional strings that are easy to transpose)
// and keyed to match the API shape, so callers can pass an event directly.
type EventTimeFields = Pick<
  EventListItem,
  'event_date' | 'start_time' | 'end_time' | 'timezone'
>

/**
 * Human-readable time range with a zone label, e.g. "3:00 PM – 5:00 PM PDT".
 * Falls back gracefully when end time or timezone is missing.
 */
export function formatEventTimeRange(event: EventTimeFields): string {
  const start = formatWallClock(event.start_time ?? '')
  if (!start) return ''
  const end = formatWallClock(event.end_time ?? '')
  const abbr = getZoneAbbreviation(event.event_date, event.timezone ?? '')
  const range = end ? `${start} – ${end}` : start
  return abbr ? `${range} ${abbr}` : range
}

/** Parses "HH:MM" or "HH:MM:SS" into minutes since midnight; null if invalid. */
function toMinutesSinceMidnight(time: string): number | null {
  const parsed = parseWallClock(time)
  if (!parsed) return null
  return parsed.hours * 60 + parsed.minutes
}

/**
 * Current wall-clock minutes-since-midnight in the given timezone. Falls back
 * to the local zone if `timezone` is invalid, so it can't throw during render.
 */
function nowMinutesInTimezone(timezone: string, now: Date): number {
  try {
    const parts = new Intl.DateTimeFormat('en-GB', {
      timeZone: timezone,
      hour: '2-digit',
      minute: '2-digit',
      hourCycle: 'h23',
    }).formatToParts(now)
    const hour = Number(parts.find((p) => p.type === 'hour')?.value ?? '0')
    const minute = Number(parts.find((p) => p.type === 'minute')?.value ?? '0')
    return hour * 60 + minute
  } catch {
    return now.getHours() * 60 + now.getMinutes()
  }
}

/**
 * Whether `now` falls within an event's [start, end) window, evaluated in the
 * event's own timezone. Events without an end time use a 1-hour default
 * duration. Requires a start time and the event being on its own local "today";
 * returns false otherwise.
 */
export function isEventHappeningNow(
  dateStr: string,
  startTime: string | null | undefined,
  endTime: string | null | undefined,
  timezone: string | null | undefined,
  now: Date = new Date(),
): boolean {
  const start = startTime ? toMinutesSinceMidnight(startTime) : null
  if (start === null) return false
  const zone = timezone || getBrowserTimezone()
  if (todayInTimezone(zone, now) !== dateStr) return false
  // Default to a 1-hour window when no end time is set.
  const end = (endTime ? toMinutesSinceMidnight(endTime) : null) ?? start + 60
  const current = nowMinutesInTimezone(zone, now)
  return current >= start && current < end
}
