import type {
  ApiDateRangeFilter,
  ApiIntRangeFilter,
  QueryActivistOptions,
} from '@/lib/api'
import type {
  DateRangeBoundValue,
  IncludeExcludeFilterValue,
  ProspectFilterValue,
} from './filter-types'
import type { FilterState } from './query-state'

type ApiFilters = QueryActivistOptions['filters']

export type FilterApiContext = {
  chapterId: number
  userId: number
}

function formatDateToYmd(date: Date): string {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

/** Resolve a typed date bound to absolute YYYY-MM-DD for API filters. */
export function resolveDateBound(value: DateRangeBoundValue): string {
  if (value.mode === 'absolute') return value.date
  const d = new Date()
  d.setDate(d.getDate() + Math.trunc(value.daysOffset))
  return formatDateToYmd(d)
}

/**
 * Converts typed date range state into API filter fields.
 * Resolves relative day offsets (integers) to absolute YYYY-MM-DD dates.
 */
export function toApiDateRange(
  value?: FilterState['lastEvent'],
): ApiDateRangeFilter | undefined {
  if (!value) return undefined
  return {
    gte: value.gte ? resolveDateBound(value.gte) : undefined,
    lt: value.lt ? resolveDateBound(value.lt) : undefined,
    or_null: value.orNull || undefined,
  }
}

/** Converts typed int range state into numeric bounds for API filters. */
export function toApiIntRange(
  value?: FilterState['totalEvents'],
): ApiIntRangeFilter | undefined {
  const parts = value
    ? {
        gte: value.gte?.toString(),
        lt: value.lt?.toString(),
      }
    : undefined
  if (!parts) return undefined
  const gte = parts.gte ? parseInt(parts.gte, 10) : undefined
  const lt = parts.lt ? parseInt(parts.lt, 10) : undefined
  if (gte === undefined && lt === undefined) return undefined
  return { gte, lt }
}

/** Convert assignedTo URL value ("me"|"any"|id) to backend integer. */
export function toApiAssignedTo(
  value: FilterState['assignedTo'],
  userId: number,
): number | undefined {
  if (!value) return undefined
  if (value === 'me') return userId
  if (value === 'any') return -1
  const n = parseInt(value, 10)
  return isNaN(n) ? undefined : n
}

const toApiSourceOrTraining = (value?: IncludeExcludeFilterValue) =>
  value && (value.include.length > 0 || value.exclude.length > 0)
    ? {
        include: value.include.length > 0 ? value.include : undefined,
        exclude: value.exclude.length > 0 ? value.exclude : undefined,
      }
    : undefined

const toApiProspectValue = (
  value?: ProspectFilterValue,
): QueryActivistOptions['filters']['prospect'] =>
  value === 'chapterMember'
    ? 'chapter_member'
    : value === 'organizer'
      ? 'organizer'
      : undefined

export const toApiSearchAcrossChapters = (
  value: FilterState['searchAcrossChapters'],
  context: FilterApiContext,
): Partial<ApiFilters> => ({
  chapter_id: value ? 0 : context.chapterId,
})

export const toApiNameSearch = (
  value: FilterState['nameSearch'],
): Partial<ApiFilters> =>
  value ? { name: { name_contains: value } } : { name: undefined }

export const toApiIncludeHidden = (
  value: FilterState['includeHidden'],
): Partial<ApiFilters> => ({ include_hidden: value })

export const toApiLastEvent = (
  value: FilterState['lastEvent'],
): Partial<ApiFilters> => ({ last_event: toApiDateRange(value) })

export const toApiInterestDate = (
  value: FilterState['interestDate'],
): Partial<ApiFilters> => ({ interest_date: toApiDateRange(value) })

export const toApiFirstEvent = (
  value: FilterState['firstEvent'],
): Partial<ApiFilters> => ({ first_event: toApiDateRange(value) })

export const toApiTotalEvents = (
  value: FilterState['totalEvents'],
): Partial<ApiFilters> => ({ total_events: toApiIntRange(value) })

export const toApiTotalInteractions = (
  value: FilterState['totalInteractions'],
): Partial<ApiFilters> => ({ total_interactions: toApiIntRange(value) })

export const toApiActivistLevel = (
  value: FilterState['activistLevel'],
): Partial<ApiFilters> => ({ activist_level: value })

export const toApiSource = (
  value: FilterState['source'],
): Partial<ApiFilters> => {
  const source = toApiSourceOrTraining(value)
  return {
    source: source
      ? {
          contains_any: source.include,
          not_contains_any: source.exclude,
        }
      : undefined,
  }
}

export const toApiTraining = (
  value: FilterState['training'],
): Partial<ApiFilters> => {
  const training = toApiSourceOrTraining(value)
  return {
    training: training
      ? {
          completed: training.include,
          not_completed: training.exclude,
        }
      : undefined,
  }
}

export const toApiAssignedToFilter = (
  value: FilterState['assignedTo'],
  context: FilterApiContext,
): Partial<ApiFilters> => ({
  assigned_to: toApiAssignedTo(value, context.userId) ?? undefined,
})

export const toApiFollowups = (
  value: FilterState['followups'],
): Partial<ApiFilters> => ({ followups: value })

export const toApiProspect = (
  value: FilterState['prospect'],
): Partial<ApiFilters> => ({ prospect: toApiProspectValue(value) })
