import {
  createLoader,
  createParser,
  type inferParserType,
  type ParserMap,
} from 'nuqs/server'
import type { ActivistColumnName } from '@/lib/api'
import {
  COLUMN_DEFINITIONS,
  DEFAULT_COLUMNS,
  normalizeColumns,
  normalizeColumnsForFilters,
} from './column-definitions'
import { buildSortParam } from './filter-url-state'
import { FILTER_NUQS_PARSERS } from './filter-nuqs-parsers'
import { FILTER_PARAM_KEYS, normalizeFilterState } from './filter-schema'
import type { ActivistsQueryState, SortColumn } from './query-state'

const VALID_COLUMN_NAMES = new Set<ActivistColumnName>(
  COLUMN_DEFINITIONS.map((column) => column.name),
)

const parseAsColumns = createParser<ActivistColumnName[]>({
  parse: (raw) => {
    const cols = normalizeColumns(
      raw
        .split(',')
        .filter((columnName): columnName is ActivistColumnName =>
          VALID_COLUMN_NAMES.has(columnName as ActivistColumnName),
        ),
    )
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
        if (!VALID_COLUMN_NAMES.has(columnName as ActivistColumnName)) {
          return []
        }
        return [{ column: columnName as ActivistColumnName, desc }]
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
