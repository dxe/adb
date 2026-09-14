import { QueryActivistOptions } from '@/lib/api'
import { normalizeColumnsForFilters } from './column-selection'
import { DEFAULT_SORT, type ActivistsQueryState } from './query-state'
import { buildApiFiltersFromState } from './filter-schema'

export type BuildQueryOptionsInput = ActivistsQueryState & {
  chapterId: number
  userId: number
  referenceDate: Date
}

export const buildQueryOptions = ({
  filters,
  selectedColumns,
  chapterId,
  userId,
  referenceDate,
  sort = DEFAULT_SORT,
}: BuildQueryOptionsInput): QueryActivistOptions => {
  let columnsToRequest = normalizeColumnsForFilters(
    selectedColumns,
    filters.searchAcrossChapters,
  )

  if (!columnsToRequest.includes('id')) {
    columnsToRequest = ['id', ...columnsToRequest]
  }

  // Always requested, never user-selectable: the table shades hidden rows
  // regardless of which columns are shown.
  if (!columnsToRequest.includes('hidden')) {
    columnsToRequest = ['hidden', ...columnsToRequest]
  }

  return {
    shape: {
      columns: columnsToRequest,
      filters: buildApiFiltersFromState(filters, {
        chapterId,
        userId,
        referenceDate,
      }),
      sort: {
        sort_columns: (sort.length > 0 ? sort : DEFAULT_SORT).map((s) => ({
          column_name: s.column,
          desc: s.desc,
        })),
      },
    },
  }
}
