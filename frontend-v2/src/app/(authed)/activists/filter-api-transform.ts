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
import { parseSafeInteger } from '@/lib/number-utils'
import {
  addDaysToYmd,
  getTodayYmdInActivistsTimeZone,
  type LocalDateYmd,
} from './date-time'

type ApiFilters = QueryActivistOptions['filters']

export type FilterApiContext = {
  chapterId: number
  userId: number
  referenceDate: Date
}

/** Resolve a typed date bound to absolute YYYY-MM-DD for API filters. */
export function resolveDateBound(
  value: DateRangeBoundValue,
  referenceDate: Date = new Date(),
): LocalDateYmd {
  if (value.mode === 'absolute') return value.date
  return addDaysToYmd(
    getTodayYmdInActivistsTimeZone(referenceDate),
    value.daysOffset,
  )
}

/**
 * Converts typed date range state into API filter fields.
 * Resolves relative day offsets (integers) to absolute YYYY-MM-DD dates.
 */
export function toApiDateRange(
  value?: FilterState['lastEvent'],
  referenceDate: Date = new Date(),
): ApiDateRangeFilter | undefined {
  if (!value) return undefined
  const gte = value.gte ? resolveDateBound(value.gte, referenceDate) : undefined
  const lt = value.lt ? resolveDateBound(value.lt, referenceDate) : undefined
  const or_null = value.orNull || undefined
  if (gte === undefined && lt === undefined && or_null === undefined) {
    return undefined
  }
  return { gte, lt, or_null }
}

/** Converts typed int range state into numeric bounds for API filters. */
export function toApiIntRange(
  value?: FilterState['totalEvents'],
): ApiIntRangeFilter | undefined {
  if (!value) return undefined
  const { gte, lt } = value
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
  return parseSafeInteger(value)
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
  context: FilterApiContext,
): Partial<ApiFilters> => ({
  last_event: toApiDateRange(value, context.referenceDate),
})

export const toApiInterestDate = (
  value: FilterState['interestDate'],
  context: FilterApiContext,
): Partial<ApiFilters> => ({
  interest_date: toApiDateRange(value, context.referenceDate),
})

export const toApiFirstEvent = (
  value: FilterState['firstEvent'],
  context: FilterApiContext,
): Partial<ApiFilters> => ({
  first_event: toApiDateRange(value, context.referenceDate),
})

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
