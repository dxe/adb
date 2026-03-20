'use client'

import { useCallback, useMemo } from 'react'
import { useQueryStates } from 'nuqs'
import type { ActivistColumnName } from '@/lib/api'
import { DEFAULT_COLUMNS } from './column-definitions'
import {
  columnsForNewFilters,
  normalizeColumnsForFilters,
} from './column-selection'
import { isFilterStateDirty } from './filter-schema'
import type { FilterState, SortColumn } from './query-state'
import {
  ACTIVIST_QUERY_STATE_PARSERS,
  ACTIVIST_QUERY_URL_KEYS,
  getActivistQueryStateFromParams,
  type ParsedActivistQueryParams,
} from './search-params'

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

function sortOrNull(sort: SortColumn[]): SortColumn[] | null {
  return sort.length > 0 ? sort : null
}

export function useActivistQueryState() {
  const [currentParams, setParams] = useQueryStates(
    ACTIVIST_QUERY_STATE_PARSERS,
    {
      history: 'replace',
      urlKeys: ACTIVIST_QUERY_URL_KEYS,
    },
  )
  const parsedParams = currentParams as ParsedActivistQueryParams

  const { filters, selectedColumns, sort } = useMemo(
    () => getActivistQueryStateFromParams(parsedParams),
    [parsedParams],
  )

  const isDirty = useMemo(
    () =>
      isFilterStateDirty(filters) ||
      parsedParams.columns !== null ||
      parsedParams.sort !== null,
    [filters, parsedParams.columns, parsedParams.sort],
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
        sort: sortOrNull(newSort),
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
        sort: sortOrNull(newSort),
      })
    },
    [filters.searchAcrossChapters, setParams, sort],
  )

  const setSort = useCallback(
    (newSort: SortColumn[]) => {
      setParams({
        sort: sortOrNull(newSort),
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
