import { ActivistColumnName } from '@/lib/api'
import { DEFAULT_COLUMNS, normalizeColumns } from './column-definitions'
import { type SortColumn } from './query-state'
import {
  FILTER_PARAM_KEYS,
  parseFiltersFromParams,
  type FilterParamGetter,
} from './filter-schema'

export type ParamGetter = FilterParamGetter

export { FILTER_PARAM_KEYS, parseFiltersFromParams }

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

/** Builds URL param value for sort state. Returns undefined only when sort is empty. */
export const buildSortParam = (sort: SortColumn[]): string | undefined => {
  if (sort.length === 0) return undefined

  return sort.map((s) => (s.desc ? `-${s.column}` : s.column)).join(',')
}
