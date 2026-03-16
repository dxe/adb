import { Temporal } from '@js-temporal/polyfill'

export const ACTIVISTS_TIME_ZONE = 'America/Los_Angeles'
export const ACTIVISTS_DISPLAY_LOCALE = 'en-US'

export type LocalDateYmd = string

const longDateFormatter = new Intl.DateTimeFormat(ACTIVISTS_DISPLAY_LOCALE, {
  timeZone: ACTIVISTS_TIME_ZONE,
  year: 'numeric',
  month: 'short',
  day: 'numeric',
})

const shortDateFormatter = new Intl.DateTimeFormat(ACTIVISTS_DISPLAY_LOCALE, {
  timeZone: ACTIVISTS_TIME_ZONE,
  month: 'short',
  day: 'numeric',
})

function toTemporalInstant(date: Date): Temporal.Instant {
  return Temporal.Instant.fromEpochMilliseconds(date.getTime())
}

function ymdToZonedDate(dateString: LocalDateYmd, timeZone: string): Date {
  const instant = Temporal.PlainDate.from(dateString)
    .toZonedDateTime({
      timeZone,
      // Use a noon anchor to avoid rollover issues that happen when date-only
      // values are converted through midnight timestamps.
      plainTime: Temporal.PlainTime.from('12:00'),
    })
    .toInstant()
  return new Date(instant.epochMilliseconds)
}

function plainDateTimeToZonedDate(
  dateTimeString: string,
  timeZone: string,
): Date {
  const normalizedDateTime = dateTimeString.replace(' ', 'T')
  const instant = Temporal.PlainDateTime.from(normalizedDateTime)
    .toZonedDateTime(timeZone)
    .toInstant()
  return new Date(instant.epochMilliseconds)
}

function parseDateValueForActivists(dateString: string): Date | undefined {
  try {
    if (/^\d{4}-\d{2}-\d{2}$/.test(dateString)) {
      return ymdToZonedDate(dateString, ACTIVISTS_TIME_ZONE)
    }

    if (
      /^\d{4}-\d{2}-\d{2}[ T]\d{2}:\d{2}(?::\d{2}(?:\.\d{1,9})?)?$/.test(
        dateString,
      )
    ) {
      return plainDateTimeToZonedDate(dateString, ACTIVISTS_TIME_ZONE)
    }

    const instant = Temporal.Instant.from(dateString)
    return new Date(instant.epochMilliseconds)
  } catch {
    const date = new Date(dateString)
    return Number.isNaN(date.getTime()) ? undefined : date
  }
}

export function getTodayYmdInActivistsTimeZone(
  referenceDate: Date = new Date(),
): LocalDateYmd {
  return toTemporalInstant(referenceDate)
    .toZonedDateTimeISO(ACTIVISTS_TIME_ZONE)
    .toPlainDate()
    .toString()
}

export function addDaysToYmd(
  dateString: LocalDateYmd,
  days: number,
): LocalDateYmd {
  return Temporal.PlainDate.from(dateString)
    .add({ days: Math.trunc(days) })
    .toString()
}

export function getCurrentYearInActivistsTimeZone(
  referenceDate: Date = new Date(),
) {
  return toTemporalInstant(referenceDate).toZonedDateTimeISO(
    ACTIVISTS_TIME_ZONE,
  ).year
}

// Date picker components typically work in the browser's timezone.
export function ymdToDatePickerValue(dateString: LocalDateYmd): Date {
  return ymdToZonedDate(dateString, Temporal.Now.timeZoneId())
}
export function datePickerValueToYmd(date: Date): LocalDateYmd | undefined {
  return toTemporalInstant(date)
    .toZonedDateTimeISO(Temporal.Now.timeZoneId())
    .toPlainDate()
    .toString()
}

export function formatYmdForActivists(
  dateString: LocalDateYmd,
  options?: { includeYear?: boolean },
): string {
  const date = ymdToZonedDate(dateString, ACTIVISTS_TIME_ZONE)
  return options?.includeYear === false
    ? shortDateFormatter.format(date)
    : longDateFormatter.format(date)
}

export function formatDateValueForActivists(dateString: string): string {
  const date = parseDateValueForActivists(dateString)

  if (!date) return dateString

  return longDateFormatter.format(date)
}
