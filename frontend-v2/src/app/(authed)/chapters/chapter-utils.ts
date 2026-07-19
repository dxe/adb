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
  const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(value)
  if (!match) return null
  const [, year, month, day] = match
  return new Date(Number(year), Number(month) - 1, Number(day))
}

export function formatDateYmd(date: Date): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

export function isDateInLastThreeMonths(value: string): boolean {
  const date = parseDateYmd(value)
  if (date == null) return false
  const threeMonthsAgo = new Date()
  threeMonthsAgo.setMonth(threeMonthsAgo.getMonth() - 3)
  return date > threeMonthsAgo
}

// Quadrimesters: Feb-May, Jun-Sep, Oct-Jan. Returns the first day of the
// quadrimester containing today.
function currentQuadrimesterStart(now: Date): Date {
  const month = now.getMonth()
  const year = now.getFullYear()
  if (month >= 1 && month <= 4) return new Date(year, 1, 1)
  if (month >= 5 && month <= 8) return new Date(year, 5, 1)
  if (month >= 9) return new Date(year, 9, 1)
  return new Date(year - 1, 9, 1) // January -> previous Oct
}

function addMonths(date: Date, months: number): Date {
  return new Date(date.getFullYear(), date.getMonth() + months, date.getDate())
}

function addDays(date: Date, days: number): Date {
  const result = new Date(date)
  result.setDate(result.getDate() + days)
  return result
}

export function colorLastAction(value: string): StatusColor {
  const now = new Date()
  const quadStart = currentQuadrimesterStart(now)
  const prevQuadStart = addMonths(quadStart, -4)
  const quadEnd = addMonths(quadStart, 4)
  const redThreshold = addDays(addMonths(quadStart, 1), 14)
  const blackThreshold = addDays(quadEnd, -7)

  const lastAction = parseDateYmd(value)
  if (lastAction == null || lastAction < prevQuadStart) return 'black'

  const hasActionThisQuadrimester = lastAction >= quadStart
  if (hasActionThisQuadrimester) return 'green'
  if (now < redThreshold) return 'green'
  if (now < blackThreshold) return 'red'
  return 'black'
}

export function lastActionTooltip(value: string): string {
  const now = new Date()
  const quadStart = currentQuadrimesterStart(now)
  const quadEnd = addDays(addMonths(quadStart, 4), -1)
  const quadrimesterText = `Current quadrimester: ${formatDateYmd(quadStart)} to ${formatDateYmd(quadEnd)}`

  const lastAction = parseDateYmd(value)
  if (lastAction == null) {
    return `No action recorded\n${quadrimesterText}`
  }
  const daysSinceLastAction = Math.floor(
    (now.getTime() - lastAction.getTime()) / (1000 * 60 * 60 * 24),
  )
  return `${daysSinceLastAction} days since last action\n${quadrimesterText}`
}

// `LastFBSync` is stored in the DB in a timezone offset by 8 hours from UTC.
export function colorFBSyncStatus(value: string): StatusColor {
  const date = parseFBSyncDate(value)
  if (date == null) return 'gray'

  const now = new Date()
  const oneHourAgo = new Date(now.getTime() - 60 * 60 * 1000)
  const oneDayAgo = new Date(now.getTime() - 24 * 60 * 60 * 1000)

  if (date > oneHourAgo) return 'green'
  if (date > oneDayAgo) return 'yellow'
  return 'red'
}

function parseFBSyncDate(value: string): Date | null {
  if (!value) return null
  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) return null
  return new Date(parsed.getTime() + 8 * 60 * 60 * 1000)
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
  return `https://mail.google.com/mail/?view=cm&fs=1&su=${chapter.Name}&to=${emails.join(',')}`
}

export const STATUS_COLOR_CLASSES: Record<StatusColor, string> = {
  green: 'bg-emerald-50 text-emerald-700',
  yellow: 'bg-amber-50 text-amber-700',
  red: 'bg-red-50 text-destructive',
  gray: 'bg-gray-100 text-muted-foreground',
  black: 'bg-gray-800 text-white',
}
