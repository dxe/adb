import type {
  ActivistLevelValue,
  ActivistLevelFilterValue,
  DateRangeBoundValue,
  DateRangeFilterValue,
} from './filter-types'
import type { FilterState } from './query-state'

const ACTIVIST_LEVEL_SLUG_BY_VALUE: Record<ActivistLevelValue, string> = {
  Supporter: 'supporter',
  'Chapter Member': 'chapter-member',
  Organizer: 'organizer',
  'Non-Local': 'non-local',
  'Global Network Member': 'global-network-member',
}

const ACTIVIST_LEVEL_FROM_SLUG = new Map<string, ActivistLevelValue>(
  Object.entries(ACTIVIST_LEVEL_SLUG_BY_VALUE).map(([value, slug]) => [
    slug,
    value as ActivistLevelValue,
  ]),
)

function decodeActivistLevelToken(
  token: string,
): ActivistLevelValue | undefined {
  if (!token) return undefined
  const level = ACTIVIST_LEVEL_FROM_SLUG.get(token)
  if (level === undefined) {
    throw Error('invalid level: ' + token)
  }
  return level
}

function decodeDateRangeBound(value?: string): DateRangeBoundValue | undefined {
  if (!value) return undefined
  if (/^-?\d+$/.test(value)) {
    return {
      mode: 'relative',
      daysOffset: parseInt(value, 10),
    }
  }
  return {
    mode: 'absolute',
    date: value,
  }
}

function encodeDateRangeBound(value?: DateRangeBoundValue): string | undefined {
  if (!value) return undefined
  if (value.mode === 'relative') {
    return String(Math.trunc(value.daysOffset))
  }
  return value.date
}

/**
 * Syntax: [gte]..[lt][|null]
 *   "2025-01-01..2025-06-01"       -> absolute gte + lt
 *   "-180.."                       -> relative gte (within last 180 days)
 *   "30.."                         -> relative gte (from 30 days in the future)
 *   "..-360|null"                  -> relative lt (over 360 days ago) or null
 *   "null"                         -> only NULL values
 */
export function decodeDateRange(
  value?: string,
): DateRangeFilterValue | undefined {
  if (!value) return undefined
  let orNull = false
  let range = value
  if (range.endsWith('|null')) {
    orNull = true
    range = range.slice(0, -5)
  }
  if (range === 'null') return { orNull: true }
  const parts = range.split('..')
  if (parts.length !== 2) return undefined
  const gte = decodeDateRangeBound(parts[0] || undefined)
  const lt = decodeDateRangeBound(parts[1] || undefined)
  if (!gte && !lt && !orNull) return undefined
  return { gte, lt, orNull }
}

/** Build date range URL syntax: [gte]..[lt][|null]. */
export function encodeDateRange(
  gte?: DateRangeBoundValue,
  lt?: DateRangeBoundValue,
  orNull?: boolean,
): string | undefined {
  const gteToken = encodeDateRangeBound(gte)
  const ltToken = encodeDateRangeBound(lt)
  const effectiveOrNull = !!orNull && !(gte && lt)
  if (!gteToken && !ltToken && !effectiveOrNull) return undefined
  if (!gteToken && !ltToken && effectiveOrNull) return 'null'
  const range = `${gteToken || ''}..${ltToken || ''}`
  return effectiveOrNull ? `${range}|null` : range
}

/** Parse "1..4" into string parts. */
export function decodeIntRange(
  value?: string,
): { gte?: string; lt?: string } | undefined {
  if (!value) return undefined
  const parts = value.split('..')
  if (parts.length !== 2) return undefined
  const gte = parts[0] || undefined
  const lt = parts[1] || undefined
  if (gte === undefined && lt === undefined) return undefined
  return { gte, lt }
}

/** Build "1..4" from parts. */
export function encodeIntRange(gte?: string, lt?: string): string | undefined {
  if (!gte && !lt) return undefined
  return `${gte || ''}..${lt || ''}`
}

/** Parse "a,b,-c" into { include: ["a","b"], exclude: ["c"] }. */
export function decodeIncludeExclude(
  value?: string,
): { include: string[]; exclude: string[] } | undefined {
  const parsed = decodeIncludeExcludeSet(value)
  if (!parsed) return undefined
  const include = Array.from(parsed.include)
  const exclude = Array.from(parsed.exclude)
  if (include.length === 0 && exclude.length === 0) return undefined
  return { include, exclude }
}

/** Parse "a,b,-c" into include/exclude sets. */
export function decodeIncludeExcludeSet(
  value?: string,
): { include: Set<string>; exclude: Set<string> } | undefined {
  if (!value) return undefined
  const include = new Set<string>()
  const exclude = new Set<string>()
  for (const part of value.split(',')) {
    const trimmed = part.trim()
    if (!trimmed) continue
    if (trimmed.startsWith('-')) {
      exclude.add(trimmed.slice(1))
    } else {
      include.add(trimmed)
    }
  }
  if (include.size === 0 && exclude.size === 0) return undefined
  return { include, exclude }
}

/** Build "a,b,-c" from include/exclude values. */
export function encodeIncludeExclude(
  include: Iterable<string>,
  exclude: Iterable<string>,
): string | undefined {
  const parts = [
    ...Array.from(include),
    ...Array.from(exclude).map((v) => `-${v}`),
  ]
  return parts.length > 0 ? parts.join(',') : undefined
}

/** Parse "a,b" or "~a,b" into { mode, values }. */
export function decodeActivistLevel(
  value?: string,
): ActivistLevelFilterValue | undefined {
  if (!value) return undefined
  const payload = value
  const mode = payload.startsWith('~') ? 'exclude' : 'include'
  const rawValues = mode === 'exclude' ? payload.slice(1) : payload
  const values = rawValues
    .split(',')
    .map((v) => v.trim())
    .map(decodeActivistLevelToken)
    .filter((v): v is ActivistLevelValue => !!v)
  return values.length > 0 ? { mode, values } : undefined
}

/** Build "a,b" or "~a,b" from mode + values. */
export function encodeActivistLevel(
  value?: ActivistLevelFilterValue,
): string | undefined {
  if (!value || value.values.length === 0) return undefined
  const encodedValues = value.values.map(
    (level) => ACTIVIST_LEVEL_SLUG_BY_VALUE[level],
  )
  return value.mode === 'exclude'
    ? `~${encodedValues.join(',')}`
    : encodedValues.join(',')
}

function parseIntValue(raw?: string): number | undefined {
  if (!raw) return undefined
  const n = parseInt(raw, 10)
  return isNaN(n) ? undefined : n
}

export const parseSearchAcrossChaptersParam = (
  raw: string | undefined,
): FilterState['searchAcrossChapters'] => raw === 'true'

export const serializeSearchAcrossChaptersParam = (
  value: FilterState['searchAcrossChapters'],
): string | undefined => (value ? 'true' : undefined)

export const parseNameSearchParam = (
  raw: string | undefined,
): FilterState['nameSearch'] => raw || ''

export const serializeNameSearchParam = (
  value: FilterState['nameSearch'],
): string | undefined => value || undefined

export const parseIncludeHiddenParam = (
  raw: string | undefined,
): FilterState['includeHidden'] => raw === 'true'

export const serializeIncludeHiddenParam = (
  value: FilterState['includeHidden'],
): string | undefined => (value ? 'true' : undefined)

export const parseDateRangeParam = (
  value?: string,
): DateRangeFilterValue | undefined => {
  const parsed = decodeDateRange(value)
  if (!parsed) return undefined
  return { ...parsed, orNull: parsed.orNull || undefined }
}

export const serializeDateRangeParam = (
  value?: FilterState['lastEvent'],
): string | undefined => encodeDateRange(value?.gte, value?.lt, value?.orNull)

export const parseIntRangeParam = (
  value?: string,
): FilterState['totalEvents'] => {
  const parsed = decodeIntRange(value)
  if (!parsed) return undefined
  const gte = parseIntValue(parsed.gte)
  const lt = parseIntValue(parsed.lt)
  if (gte === undefined && lt === undefined) return undefined
  return { gte, lt }
}

export const serializeIntRangeParam = (
  value?: FilterState['totalEvents'],
): string | undefined =>
  encodeIntRange(value?.gte?.toString(), value?.lt?.toString())

export const parseActivistLevelParam = (
  value?: string,
): FilterState['activistLevel'] => decodeActivistLevel(value)

export const serializeActivistLevelParam = (
  value?: FilterState['activistLevel'],
): string | undefined => encodeActivistLevel(value)

export const parseIncludeExcludeParam = (
  value?: string,
): FilterState['source'] => {
  const parsed = decodeIncludeExclude(value)
  if (!parsed) return undefined
  return {
    include: parsed.include,
    exclude: parsed.exclude,
  }
}

export const serializeIncludeExcludeParam = (
  value?: FilterState['source'],
): string | undefined =>
  encodeIncludeExclude(value?.include ?? [], value?.exclude ?? [])

export const parseAssignedToParam = (
  value?: string,
): FilterState['assignedTo'] => {
  if (!value) return undefined
  if (value === 'me' || value === 'any') return value
  return /^-?\d+$/.test(value) ? (value as `${number}`) : undefined
}

export const serializeAssignedToParam = (
  value?: FilterState['assignedTo'],
): string | undefined => value

export const parseFollowupsParam = (
  value?: string,
): FilterState['followups'] => {
  if (value === 'all' || value === 'due' || value === 'upcoming') {
    return value
  }
  return undefined
}

export const serializeFollowupsParam = (
  value?: FilterState['followups'],
): string | undefined => value

export const parseProspectParam = (value?: string): FilterState['prospect'] => {
  if (value === 'chapterMember' || value === 'organizer') {
    return value
  }
  return undefined
}

export const serializeProspectParam = (
  value?: FilterState['prospect'],
): string | undefined => value
