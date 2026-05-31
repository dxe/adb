// Timezone helpers for advance events.
//
// Future-dated events store a local calendar date, local wall-clock times, and
// an IANA timezone name (e.g. "America/Los_Angeles"). Display always formats in
// the event's own timezone with a label, correct for any viewer. We rely on the
// built-in Intl APIs rather than an extra date library.

// A small, ordered curated list of common zones. Used for manual selection
// (the full ~400-zone IANA set is overwhelming and we don't need it here) and
// as a fallback when the runtime can't enumerate zones.
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

// The creator's browser timezone, used as the default on create.
export function getBrowserTimezone(): string {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC'
  } catch {
    return 'UTC'
  }
}

// The curated short list for manual timezone selection, always including the
// browser zone so it is selectable. The full IANA enumeration is intentionally
// not offered — these common zones cover where the org operates.
export function getCommonTimezones(): string[] {
  const browser = getBrowserTimezone()
  return COMMON_TIMEZONES.includes(browser)
    ? COMMON_TIMEZONES
    : [browser, ...COMMON_TIMEZONES]
}

// Today's calendar date (YYYY-MM-DD) in the given timezone. Used to compute
// "today's events" in a defined zone rather than the DB/browser raw date.
export function todayInTimezone(timezone: string): string {
  // en-CA formats as YYYY-MM-DD.
  return new Intl.DateTimeFormat('en-CA', {
    timeZone: timezone,
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  }).format(new Date())
}

// Short timezone abbreviation (e.g. "PDT") for the given date in the given zone.
// We pick noon UTC on the event's date so the abbreviation lands on the correct
// side of any DST transition for essentially all real-world cases.
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

// Formats a 24h "HH:MM" wall-clock string as a 12h label (e.g. "3:00 PM").
// Returns '' for empty/invalid input.
export function formatWallClock(time: string): string {
  const match = /^(\d{1,2}):(\d{2})/.exec(time.trim())
  if (!match) return ''
  const hours = Number(match[1])
  const minutes = match[2]
  const period = hours >= 12 ? 'PM' : 'AM'
  const hour12 = hours % 12 === 0 ? 12 : hours % 12
  return `${hour12}:${minutes} ${period}`
}

// Human-readable time range with a zone label, e.g. "3:00 PM – 5:00 PM PDT".
// Falls back gracefully when end time or timezone is missing.
export function formatEventTimeRange(
  dateStr: string,
  startTime: string,
  endTime: string,
  timezone: string,
): string {
  const start = formatWallClock(startTime)
  if (!start) return ''
  const end = formatWallClock(endTime)
  const abbr = getZoneAbbreviation(dateStr, timezone)
  const range = end ? `${start} – ${end}` : start
  return abbr ? `${range} ${abbr}` : range
}

// Parses "HH:MM" or "HH:MM:SS" into minutes since midnight; null if invalid.
function toMinutesSinceMidnight(time: string): number | null {
  const match = /^(\d{1,2}):(\d{2})/.exec(time.trim())
  if (!match) return null
  return Number(match[1]) * 60 + Number(match[2])
}

// Current wall-clock minutes-since-midnight in the given timezone.
function nowMinutesInTimezone(timezone: string, now: Date): number {
  const parts = new Intl.DateTimeFormat('en-GB', {
    timeZone: timezone,
    hour: '2-digit',
    minute: '2-digit',
    hourCycle: 'h23',
  }).formatToParts(now)
  const hour = Number(parts.find((p) => p.type === 'hour')?.value ?? '0')
  const minute = Number(parts.find((p) => p.type === 'minute')?.value ?? '0')
  return hour * 60 + minute
}

// Whether `now` falls within an event's [start, end) window, evaluated in the
// event's own timezone. Events without an end time use a 1-hour default
// duration. Requires a start time and the event being on its own local "today";
// returns false otherwise.
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
  if (todayInTimezone(zone) !== dateStr) return false
  // Default to a 1-hour window when no end time is set.
  const end = (endTime ? toMinutesSinceMidnight(endTime) : null) ?? start + 60
  const current = nowMinutesInTimezone(zone, now)
  return current >= start && current < end
}
