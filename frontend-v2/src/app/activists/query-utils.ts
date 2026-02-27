import { ActivistColumnName, QueryActivistOptions } from '@/lib/api'
import {
  DEFAULT_COLUMNS,
  normalizeColumns,
  normalizeColumnsForFilters,
} from './column-definitions'

export type FilterState = {
  searchAcrossChapters: boolean
  nameSearch: string
  includeHidden: boolean
  // Range filters stored as URL-format strings (e.g. "2025-01-01..2025-06-01|null")
  lastEvent?: string
  interestDate?: string
  firstEvent?: string
  totalEvents?: string
  totalInteractions?: string
  // Include/exclude filters stored as URL-format strings (e.g. "Supporter,-Organizer")
  activistLevel?: string
  source?: string
  training?: string
  // Simple value filters
  assignedTo?: string // "me" | "any" | numeric string
  followups?: string // "all" | "due" | "upcoming"
  prospect?: string // "chapterMember" | "organizer"
}

export type SortColumn = {
  column: ActivistColumnName
  desc: boolean
}

export const DEFAULT_SORT: SortColumn[] = [{ column: 'name', desc: false }]

export type ParamGetter = (key: string) => string | undefined

export const parseFiltersFromParams = (getParam: ParamGetter): FilterState => ({
  searchAcrossChapters: getParam('searchAcrossChapters') === 'true',
  nameSearch: getParam('nameSearch') || '',
  includeHidden: getParam('includeHidden') === 'true',
  lastEvent: getParam('lastEvent') || undefined,
  interestDate: getParam('interestDate') || undefined,
  firstEvent: getParam('firstEvent') || undefined,
  totalEvents: getParam('totalEvents') || undefined,
  totalInteractions: getParam('totalInteractions') || undefined,
  activistLevel: getParam('activistLevel') || undefined,
  source: getParam('source') || undefined,
  training: getParam('training') || undefined,
  assignedTo: getParam('assignedTo') || undefined,
  followups: getParam('followups') || undefined,
  prospect: getParam('prospect') || undefined,
})

export const parseColumnsFromParams = (
  getParam: ParamGetter,
): ActivistColumnName[] => {
  const columnsParam = getParam('columns') || ''
  return columnsParam
    ? normalizeColumns(columnsParam.split(',') as ActivistColumnName[])
    : DEFAULT_COLUMNS
}

export const parseSortFromParams = (
  getParam: ParamGetter,
  visibleColumns: ActivistColumnName[],
): SortColumn[] => {
  const sortParam = getParam('sort')
  if (!sortParam) return []

  return sortParam
    .split(',')
    .slice(0, 2)
    .map((part) => {
      const desc = part.startsWith('-')
      const column = (desc ? part.slice(1) : part) as ActivistColumnName
      return { column, desc }
    })
    .filter((s) => visibleColumns.includes(s.column))
}

/** Builds URL param value for sort state. Returns undefined for default/empty sort. */
export const buildSortParam = (sort: SortColumn[]): string | undefined => {
  if (sort.length === 0) return undefined
  const isDefault =
    sort.length === DEFAULT_SORT.length &&
    sort.every(
      (s, i) =>
        s.column === DEFAULT_SORT[i].column && s.desc === DEFAULT_SORT[i].desc,
    )
  if (isDefault) return undefined

  return sort.map((s) => (s.desc ? `-${s.column}` : s.column)).join(',')
}

// --- URL range syntax helpers ---

/**
 * Parses the URL date range syntax into API filter fields.
 *
 * Syntax: [gte]..[lt][|null]
 *   "2025-01-01..2025-06-01"       → gte + lt (between two dates)
 *   "2025-01-01.."                  → gte only (on or after)
 *   "..2025-06-01"                  → lt only (before)
 *   "..2025-06-01|null"             → lt, including NULL values
 *   "2025-01-01..|null"             → gte, including NULL values
 *   "2025-01-01..2025-06-01|null"   → gte + lt, including NULL values
 *   "null"                          → only NULL values
 */
function parseDateRange(
  value?: string,
): { gte?: string; lt?: string; orNull?: boolean } | undefined {
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
  const gte = parts[0] || undefined
  const lt = parts[1] || undefined
  if (!gte && !lt && !orNull) return undefined
  return { gte, lt, orNull: orNull || undefined }
}

/** Parse "1..4" into {gte, lt}. */
function parseIntRange(
  value?: string,
): { gte?: number; lt?: number } | undefined {
  if (!value) return undefined
  const parts = value.split('..')
  if (parts.length !== 2) return undefined
  const gte = parts[0] ? parseInt(parts[0], 10) : undefined
  const lt = parts[1] ? parseInt(parts[1], 10) : undefined
  if (gte === undefined && lt === undefined) return undefined
  return { gte, lt }
}

/** Parse "a,b,-c" into { include: ["a","b"], exclude: ["c"] }. */
function parseIncludeExclude(
  value?: string,
): { include: string[]; exclude: string[] } | undefined {
  if (!value) return undefined
  const include: string[] = []
  const exclude: string[] = []
  for (const part of value.split(',')) {
    const trimmed = part.trim()
    if (!trimmed) continue
    if (trimmed.startsWith('-')) {
      exclude.push(trimmed.slice(1))
    } else {
      include.push(trimmed)
    }
  }
  if (include.length === 0 && exclude.length === 0) return undefined
  return { include, exclude }
}

/** Convert assignedTo URL value ("me"|"any"|id) to backend integer. */
function parseAssignedTo(value: string | undefined, userId: number): number | undefined {
  if (!value) return undefined
  if (value === 'me') return userId
  if (value === 'any') return -1
  const n = parseInt(value, 10)
  return isNaN(n) ? undefined : n
}

// --- Build API query options ---

export const buildQueryOptions = ({
  filters,
  selectedColumns,
  chapterId,
  userId,
  nameSearch = filters.nameSearch,
  sort = DEFAULT_SORT,
}: {
  filters: FilterState
  selectedColumns: ActivistColumnName[]
  chapterId: number
  userId: number
  nameSearch?: string
  sort?: SortColumn[]
}): QueryActivistOptions => {
  let columnsToRequest = normalizeColumnsForFilters(
    selectedColumns,
    filters.searchAcrossChapters,
  )

  if (!columnsToRequest.includes('id')) {
    columnsToRequest = ['id', ...columnsToRequest]
  }

  const lastEvent = parseDateRange(filters.lastEvent)
  const interestDate = parseDateRange(filters.interestDate)
  const firstEvent = parseDateRange(filters.firstEvent)
  const totalEvents = parseIntRange(filters.totalEvents)
  const totalInteractions = parseIntRange(filters.totalInteractions)
  const activistLevel = parseIncludeExclude(filters.activistLevel)
  const source = parseIncludeExclude(filters.source)
  const training = parseIncludeExclude(filters.training)

  const assignedTo = parseAssignedTo(filters.assignedTo, userId)

  return {
    columns: columnsToRequest,
    filters: {
      chapter_id: filters.searchAcrossChapters ? 0 : chapterId,
      name: nameSearch ? { name_contains: nameSearch } : undefined,
      last_event: lastEvent
        ? { gte: lastEvent.gte, lt: lastEvent.lt, or_null: lastEvent.orNull }
        : undefined,
      include_hidden: filters.includeHidden,
      activist_level:
        activistLevel &&
        (activistLevel.include.length > 0 ||
          activistLevel.exclude.length > 0)
          ? {
              include:
                activistLevel.include.length > 0
                  ? activistLevel.include
                  : undefined,
              exclude:
                activistLevel.exclude.length > 0
                  ? activistLevel.exclude
                  : undefined,
            }
          : undefined,
      interest_date: interestDate
        ? { gte: interestDate.gte, lt: interestDate.lt, or_null: interestDate.orNull }
        : undefined,
      first_event: firstEvent
        ? { gte: firstEvent.gte, lt: firstEvent.lt, or_null: firstEvent.orNull }
        : undefined,
      total_events: totalEvents
        ? { gte: totalEvents.gte, lt: totalEvents.lt }
        : undefined,
      total_interactions: totalInteractions
        ? { gte: totalInteractions.gte, lt: totalInteractions.lt }
        : undefined,
      source:
        source &&
        (source.include.length > 0 || source.exclude.length > 0)
          ? {
              contains_any:
                source.include.length > 0 ? source.include : undefined,
              not_contains_any:
                source.exclude.length > 0 ? source.exclude : undefined,
            }
          : undefined,
      training:
        training &&
        (training.include.length > 0 || training.exclude.length > 0)
          ? {
              completed:
                training.include.length > 0 ? training.include : undefined,
              not_completed:
                training.exclude.length > 0 ? training.exclude : undefined,
            }
          : undefined,
      assigned_to: assignedTo ?? undefined,
      followups: filters.followups as
        | 'all'
        | 'due'
        | 'upcoming'
        | undefined,
      prospect:
        filters.prospect === 'chapterMember'
          ? 'chapter_member'
          : filters.prospect === 'organizer'
            ? 'organizer'
            : undefined,
    },
    sort: {
      sort_columns: (sort.length > 0 ? sort : DEFAULT_SORT).map((s) => ({
        column_name: s.column,
        desc: s.desc,
      })),
    },
  }
}
