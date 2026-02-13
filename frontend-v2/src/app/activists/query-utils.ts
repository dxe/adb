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

export const buildQueryOptions = ({
  filters,
  selectedColumns,
  chapterId,
  nameSearch = filters.nameSearch,
}: {
  filters: FilterState
  selectedColumns: ActivistColumnName[]
  chapterId: number
  nameSearch?: string
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
  }
}
