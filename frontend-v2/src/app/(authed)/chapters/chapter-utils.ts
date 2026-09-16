import {
  addDays,
  addHours,
  addMonths,
  differenceInCalendarDays,
  format,
  isAfter,
  isBefore,
  isValid,
  parse,
  subDays,
  subHours,
  subMonths,
} from 'date-fns'

export const REGIONS = [
  'North America',
  'Central & South America',
  'Europe',
  'Middle East & Africa',
  'Asia-Pacific',
  'Online',
] as const

export type StatusColor = 'green' | 'yellow' | 'red' | 'gray' | 'black'

/** Parses a `YYYY-MM-DD` string as a local calendar date (no timezone shift). */
export function parseDateYmd(value: string): Date | null {
  const parsed = parse(value, 'yyyy-MM-dd', new Date())
  return isValid(parsed) ? parsed : null
}

export function formatDateYmd(date: Date): string {
  return format(date, 'yyyy-MM-dd')
}

export function isDateInLastThreeMonths(value: string): boolean {
  const date = parseDateYmd(value)
  if (date == null) return false
  return isAfter(date, subMonths(new Date(), 3))
}

// Quadrimesters: Feb-May, Jun-Sep, Oct-Jan. Returns the first day of the
// quadrimester containing today. Ported from frontend/ChapterList.vue.
function currentQuadrimesterStart(now: Date): Date {
  const month = now.getMonth()
  const year = now.getFullYear()
  if (month >= 1 && month <= 4) return new Date(year, 1, 1)
  if (month >= 5 && month <= 8) return new Date(year, 5, 1)
  if (month >= 9) return new Date(year, 9, 1)
  return new Date(year - 1, 9, 1) // January -> previous Oct
}

export function colorLastAction(value: string): StatusColor {
  const now = new Date()
  const quadStart = currentQuadrimesterStart(now)
  const prevQuadStart = subMonths(quadStart, 4)
  const quadEnd = addMonths(quadStart, 4)
  const redThreshold = addDays(addMonths(quadStart, 1), 14)
  const blackThreshold = subDays(quadEnd, 7)

  const lastAction = parseDateYmd(value)
  if (lastAction == null || isBefore(lastAction, prevQuadStart)) return 'black'

  const hasActionThisQuadrimester = !isBefore(lastAction, quadStart)
  if (hasActionThisQuadrimester) return 'green'
  if (isBefore(now, redThreshold)) return 'green'
  if (isBefore(now, blackThreshold)) return 'red'
  return 'black'
}

export function lastActionTooltip(value: string): string {
  const now = new Date()
  const quadStart = currentQuadrimesterStart(now)
  const quadEnd = subDays(addMonths(quadStart, 4), 1)
  const quadrimesterText = `Current quadrimester: ${formatDateYmd(quadStart)} to ${formatDateYmd(quadEnd)}`

  const lastAction = parseDateYmd(value)
  if (lastAction == null) {
    return `No action recorded\n${quadrimesterText}`
  }
  const daysSinceLastAction = differenceInCalendarDays(now, lastAction)
  return `${daysSinceLastAction} days since last action\n${quadrimesterText}`
}

// `LastFBSync` is stored in the DB in a timezone offset by 8 hours from UTC.
export function colorFBSyncStatus(value: string): StatusColor {
  const date = parseFBSyncDate(value)
  if (date == null) return 'gray'

  const now = new Date()
  if (isAfter(date, subHours(now, 1))) return 'green'
  if (isAfter(date, subDays(now, 1))) return 'yellow'
  return 'red'
}

function parseFBSyncDate(value: string): Date | null {
  if (!value) return null
  const parsed = new Date(value)
  return isValid(parsed) ? addHours(parsed, 8) : null
}

/** Builds a Gmail compose link addressed to a chapter's own email plus all of its organizers' emails. */
export function buildChapterEmailLink(chapter: {
  Name: string
  Email: string
  Organizers: { Email: string }[]
}): string | null {
  const emails = [
    chapter.Email,
    ...chapter.Organizers.map((o) => o.Email),
  ].filter(Boolean)
  if (emails.length === 0) return null
  const params = new URLSearchParams({
    view: 'cm',
    fs: '1',
    su: chapter.Name,
    to: emails.join(','),
  })
  return `https://mail.google.com/mail/?${params}`
}

export const STATUS_COLOR_CLASSES: Record<StatusColor, string> = {
  green: 'bg-emerald-50 text-emerald-700',
  yellow: 'bg-amber-50 text-amber-700',
  red: 'bg-red-50 text-destructive',
  gray: 'bg-gray-100 text-muted-foreground',
  black: 'bg-gray-800 text-white',
}
