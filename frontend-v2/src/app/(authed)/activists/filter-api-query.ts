import { ActivistColumnName, QueryActivistOptions } from '@/lib/api'
import { normalizeColumnsForFilters } from './column-definitions'
import { DEFAULT_SORT, type FilterState, type SortColumn } from './query-state'
import { buildApiFiltersFromState } from './filter-schema'

export const buildQueryOptions = ({
  filters,
  selectedColumns,
  chapterId,
  userId,
  referenceDate,
  sort = DEFAULT_SORT,
}: {
  filters: FilterState
  selectedColumns: ActivistColumnName[]
  chapterId: number
  userId: number
  referenceDate: Date
  sort?: SortColumn[]
}): QueryActivistOptions => {
  let columnsToRequest = normalizeColumnsForFilters(
    selectedColumns,
    filters.searchAcrossChapters,
  )

  if (!columnsToRequest.includes('id')) {
    columnsToRequest = ['id', ...columnsToRequest]
  }

  return {
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
  }
}
