import { QueryActivistOptions, type ActivistColumnName } from '@/lib/api'
import { normalizeColumnsForFilters } from './column-selection'
import { DEFAULT_SORT, type ActivistsQueryState } from './query-state'
import { buildApiFiltersFromState } from './filter-schema'

export type BuildQueryOptionsInput = ActivistsQueryState & {
  chapterId: number
  userId: number
  referenceDate: Date
}

/**
 * Narrows query options to an explicit set of activists, for acting on the
 * rows the user selected rather than on everything the filters match.
 *
 * The other filters are dropped rather than combined with the ids: the user
 * picked these rows, so a row that has since stopped matching a filter (see
 * the stale-results notice) still belongs in the result. `chapter_id` stays —
 * it is what the server authorizes the query against — as does
 * `include_hidden`, so this can't reach a hidden activist that the query the
 * selection was made in wouldn't return.
 */
export const buildSelectionQueryOptions = (
  queryOptions: QueryActivistOptions,
  activistIds: number[],
): QueryActivistOptions => ({
  ...queryOptions,
  shape: {
    ...queryOptions.shape,
    filters: {
      chapter_id: queryOptions.shape.filters.chapter_id,
      include_hidden: queryOptions.shape.filters.include_hidden,
      ids: activistIds,
    },
  },
})

/**
 * Replaces the requested columns with the ones the table shows.
 *
 * buildQueryOptions also asks for `id` and `hidden` because the table needs
 * them to identify and shade its rows, but neither is a column the user chose
 * to see, so an export shouldn't carry them.
 */
export const withVisibleColumns = (
  queryOptions: QueryActivistOptions,
  visibleColumns: ActivistColumnName[],
): QueryActivistOptions => ({
  ...queryOptions,
  shape: { ...queryOptions.shape, columns: visibleColumns },
})

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
