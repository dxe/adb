import {
  createLoader,
  createParser,
  type inferParserType,
  type ParserMap,
} from 'nuqs/server'
import type { ActivistColumnName } from '@/lib/api'
import { DEFAULT_COLUMNS, isActivistColumnName } from './column-definitions'
import { buildSortParam } from './filter-url-state'
import { FILTER_NUQS_PARSERS } from './filter-nuqs-parsers'
import { FILTER_PARAM_KEYS, normalizeFilterState } from './filter-schema'
import {
  normalizeColumns,
  normalizeColumnsForFilters,
} from './column-selection'
import type { ActivistsQueryState, SortColumn } from './query-state'

const parseAsColumns = createParser<ActivistColumnName[]>({
  parse: (raw) => {
    const cols = normalizeColumns(raw.split(',').filter(isActivistColumnName))
    return cols.length > 0 ? cols : null
  },
  serialize: (cols) =>
    cols.filter((c) => c !== 'chapter_name' && c !== 'name').join(','),
})

const parseAsSort = createParser<SortColumn[]>({
  parse: (raw) => {
    const parts = raw
      .split(',')
      .slice(0, 2)
      .flatMap((part) => {
        const desc = part.startsWith('-')
        const columnName = desc ? part.slice(1) : part
        if (!isActivistColumnName(columnName)) {
          return []
        }
        return [{ column: columnName, desc }]
      })
    return parts.length > 0 ? parts : null
  },
  serialize: (sort) => buildSortParam(sort) ?? '',
})

export const ACTIVIST_QUERY_STATE_PARSERS = {
  ...FILTER_NUQS_PARSERS,
  columns: parseAsColumns,
  sort: parseAsSort,
} satisfies ParserMap

export const ACTIVIST_QUERY_URL_KEYS = {
  ...FILTER_PARAM_KEYS,
  columns: 'columns',
  sort: 'sort',
} as const

export const loadActivistSearchParams = createLoader(
  ACTIVIST_QUERY_STATE_PARSERS,
  {
    urlKeys: ACTIVIST_QUERY_URL_KEYS,
  },
)

export type ParsedActivistQueryParams = inferParserType<
  typeof ACTIVIST_QUERY_STATE_PARSERS
>

export function getActivistQueryStateFromParams(
  params: ParsedActivistQueryParams,
): ActivistsQueryState {
  const filters = normalizeFilterState(params)
  const selectedColumns = normalizeColumnsForFilters(
    params.columns ?? DEFAULT_COLUMNS,
    filters.searchAcrossChapters,
  )
  const sort = (params.sort ?? []).filter((s) =>
    selectedColumns.includes(s.column),
  )

  return {
    filters,
    selectedColumns,
    sort,
  }
}
