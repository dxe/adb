'use client'

import { useCallback, useMemo } from 'react'
import { createParser, useQueryStates } from 'nuqs'
import type { ActivistColumnName } from '@/lib/api'
import {
  DEFAULT_COLUMNS,
  columnsForNewFilters,
  normalizeColumns,
  normalizeColumnsForFilters,
} from './column-definitions'
import { buildSortParam } from './filter-url-state'
import type { FilterState, SortColumn } from './query-state'
import {
  FILTER_PARAM_KEYS,
  isFilterStateDirty,
  normalizeFilterState,
} from './filter-schema'
import { FILTER_NUQS_PARSERS } from './filter-nuqs-parsers'

const parseAsColumns = createParser<ActivistColumnName[]>({
  parse: (raw) => {
    const cols = normalizeColumns(raw.split(',') as ActivistColumnName[])
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
      .map((part) => {
        const desc = part.startsWith('-')
        const column = (desc ? part.slice(1) : part) as ActivistColumnName
        return { column, desc }
      })
    return parts.length > 0 ? parts : null
  },
  serialize: (sort) => buildSortParam(sort) ?? '',
})

const activistQueryStateParsers = {
  ...FILTER_NUQS_PARSERS,
  columns: parseAsColumns,
  sort: parseAsSort,
}

/** Returns null when columns match the defaults (removes the param). */
function columnsToStoreParam(
  normalized: ActivistColumnName[],
  searchAcrossChapters: boolean,
): ActivistColumnName[] | null {
  const defaultCols = normalizeColumnsForFilters(
    DEFAULT_COLUMNS,
    searchAcrossChapters,
  )
  const match =
    normalized.length === defaultCols.length &&
    normalized.every((c, i) => c === defaultCols[i])
  return match ? null : normalized
}

export function useActivistQueryState() {
  const [params, setParams] = useQueryStates(activistQueryStateParsers, {
    history: 'replace',
    urlKeys: FILTER_PARAM_KEYS,
  })

  const filters: FilterState = useMemo(
    () => normalizeFilterState(params),
    [params],
  )

  const selectedColumns = useMemo(
    () =>
      normalizeColumnsForFilters(
        params.columns ?? DEFAULT_COLUMNS,
        filters.searchAcrossChapters,
      ),
    [filters.searchAcrossChapters, params.columns],
  )

  const sort = useMemo(
    () => (params.sort ?? []).filter((s) => selectedColumns.includes(s.column)),
    [params.sort, selectedColumns],
  )

  const isDirty = useMemo(
    () =>
      isFilterStateDirty(filters) ||
      params.columns !== null ||
      params.sort !== null,
    [filters, params.columns, params.sort],
  )

  const setFilters = useCallback(
    (newFilters: FilterState) => {
      const autoColumns = columnsForNewFilters(filters, newFilters)
      const withAutoColumns =
        autoColumns.length > 0
          ? [...selectedColumns, ...autoColumns]
          : selectedColumns
      const newSelectedColumns = normalizeColumnsForFilters(
        withAutoColumns,
        newFilters.searchAcrossChapters,
      )
      const newSort = sort.every((entry) =>
        newSelectedColumns.includes(entry.column),
      )
        ? sort
        : []

      setParams({
        searchAcrossChapters: newFilters.searchAcrossChapters,
        nameSearch: newFilters.nameSearch,
        includeHidden: newFilters.includeHidden,
        lastEvent: newFilters.lastEvent ?? null,
        interestDate: newFilters.interestDate ?? null,
        firstEvent: newFilters.firstEvent ?? null,
        totalEvents: newFilters.totalEvents ?? null,
        totalInteractions: newFilters.totalInteractions ?? null,
        activistLevel: newFilters.activistLevel ?? null,
        source: newFilters.source ?? null,
        training: newFilters.training ?? null,
        assignedTo: newFilters.assignedTo ?? null,
        followups: newFilters.followups ?? null,
        prospect: newFilters.prospect ?? null,
        columns: columnsToStoreParam(
          newSelectedColumns,
          newFilters.searchAcrossChapters,
        ),
        sort: buildSortParam(newSort) !== undefined ? newSort : null,
      })
    },
    [filters, selectedColumns, setParams, sort],
  )

  const setSelectedColumns = useCallback(
    (columns: ActivistColumnName[]) => {
      const normalized = normalizeColumnsForFilters(
        columns,
        filters.searchAcrossChapters,
      )
      const newSort = sort.every((s) => normalized.includes(s.column))
        ? sort
        : []
      setParams({
        columns: columnsToStoreParam(normalized, filters.searchAcrossChapters),
        sort: buildSortParam(newSort) !== undefined ? newSort : null,
      })
    },
    [filters.searchAcrossChapters, setParams, sort],
  )

  const setSort = useCallback(
    (newSort: SortColumn[]) => {
      setParams({
        sort: buildSortParam(newSort) !== undefined ? newSort : null,
      })
    },
    [setParams],
  )

  const resetAll = useCallback(() => setParams(null), [setParams])

  return {
    filters,
    selectedColumns,
    sort,
    isDirty,
    setFilters,
    setSelectedColumns,
    setSort,
    resetAll,
  }
}
