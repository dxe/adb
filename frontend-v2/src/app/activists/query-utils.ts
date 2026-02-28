import { ActivistColumnName, QueryActivistOptions } from '@/lib/api'
import {
  DEFAULT_COLUMNS,
  normalizeColumns,
  normalizeColumnsForFilters,
} from './column-definitions'

export type FilterState = {
  searchAcrossChapters: boolean
  nameSearch: string
  lastEventLt?: string // ISO date string
  lastEventGte?: string // ISO date string
  includeHidden: boolean
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
  lastEventGte: getParam('lastEventGte') || undefined,
  lastEventLt: getParam('lastEventLt') || undefined,
  includeHidden: getParam('includeHidden') === 'true',
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

export const buildQueryOptions = ({
  filters,
  selectedColumns,
  chapterId,
  nameSearch = filters.nameSearch,
  sort = DEFAULT_SORT,
}: {
  filters: FilterState
  selectedColumns: ActivistColumnName[]
  chapterId: number
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

  return {
    columns: columnsToRequest,
    filters: {
      chapter_id: filters.searchAcrossChapters ? 0 : chapterId,
      name: nameSearch ? { name_contains: nameSearch } : undefined,
      last_event:
        filters.lastEventGte || filters.lastEventLt
          ? {
              last_event_gte: filters.lastEventGte,
              last_event_lt: filters.lastEventLt,
            }
          : undefined,
      include_hidden: filters.includeHidden,
    },
    sort: {
      sort_columns: (sort.length > 0 ? sort : DEFAULT_SORT).map((s) => ({
        column_name: s.column,
        desc: s.desc,
      })),
    },
  }
}
