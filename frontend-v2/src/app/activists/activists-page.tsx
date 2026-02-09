'use client'

import { useMemo, useState, useEffect, useCallback, useReducer } from 'react'
import { useQuery } from '@tanstack/react-query'
import { liteDebounce } from '@tanstack/pacer-lite'
import { useSearchParams, useRouter } from 'next/navigation'
import {
  apiClient,
  API_PATH,
  QueryActivistOptions,
  ActivistColumnName,
} from '@/lib/api'
import { useAuthedPageContext } from '@/hooks/useAuthedPageContext'
import { ActivistTable } from './activists-table'
import { ActivistFilters } from './activist-filters'
import { ColumnSelector } from './column-selector'
import { normalizeColumnsForFilters } from './column-definitions'
import {
  FilterState,
  buildQueryOptions,
  parseColumnsFromParams,
  parseFiltersFromParams,
} from './query-utils'

const BASE_PATH = '/activists'

const buildUrlParams = (
  filters: FilterState,
  visibleColumns: ActivistColumnName[],
): URLSearchParams => {
  const params = new URLSearchParams()

  if (filters.searchAcrossChapters) {
    params.set('searchAcrossChapters', 'true')
  }
  if (filters.nameSearch) {
    params.set('nameSearch', filters.nameSearch)
  }
  if (filters.lastEventGte) {
    params.set('lastEventGte', filters.lastEventGte)
  }
  if (filters.lastEventLt) {
    params.set('lastEventLt', filters.lastEventLt)
  }
  if (filters.includeHidden) {
    params.set('includeHidden', 'true')
  }

  const columnsToStore = visibleColumns.filter(
    (col) => col !== 'chapter_name' && col !== 'name',
  )
  if (columnsToStore.length > 0) {
    params.set('columns', columnsToStore.join(','))
  }

  return params
}

type ActivistsState = {
  filters: FilterState
  selectedColumns: ActivistColumnName[]
}

type ActivistsAction =
  | { type: 'setFilters'; filters: FilterState }
  | { type: 'setSelectedColumns'; columns: ActivistColumnName[] }

const activistsReducer = (
  state: ActivistsState,
  action: ActivistsAction,
): ActivistsState => {
  switch (action.type) {
    case 'setFilters': {
      const { filters } = action
      const selectedColumns =
        filters.searchAcrossChapters === state.filters.searchAcrossChapters
          ? state.selectedColumns
          : normalizeColumnsForFilters(
              state.selectedColumns,
              filters.searchAcrossChapters,
            )
      return { filters, selectedColumns }
    }
    case 'setSelectedColumns': {
      return {
        filters: state.filters,
        selectedColumns: normalizeColumnsForFilters(
          action.columns,
          state.filters.searchAcrossChapters,
        ),
      }
    }
    default:
      return state
  }
}

export default function ActivistsPage() {
  const { user } = useAuthedPageContext()
  const isAdmin = user.role === 'admin'
  const searchParams = useSearchParams()
  const router = useRouter()

  const getParam = useCallback(
    (key: string) => searchParams.get(key) || undefined,
    [searchParams],
  )
  const initialFilters = parseFiltersFromParams(getParam)
  const initialColumns = parseColumnsFromParams(getParam)

  const [state, dispatch] = useReducer(activistsReducer, {
    filters: initialFilters,
    selectedColumns: normalizeColumnsForFilters(
      initialColumns,
      initialFilters.searchAcrossChapters,
    ),
  })
  const { filters, selectedColumns } = state

  const [debouncedNameSearch, setDebouncedNameSearch] = useState(
    filters.nameSearch,
  )

  const debouncedSetNameSearch = useMemo(
    () =>
      liteDebounce((value: string) => setDebouncedNameSearch(value), {
        wait: 300,
      }),
    [],
  )

  useEffect(() => {
    debouncedSetNameSearch(filters.nameSearch)
  }, [filters.nameSearch, debouncedSetNameSearch])

  // Update URL when filters or columns change
  useEffect(() => {
    const params = buildUrlParams(filters, selectedColumns).toString()
    const newUrl = params ? `?${params}` : BASE_PATH
    router.replace(newUrl, { scroll: false })
  }, [filters, selectedColumns, router])

  const handleFiltersChange = useCallback((newFilters: FilterState) => {
    dispatch({ type: 'setFilters', filters: newFilters })
  }, [])

  const handleColumnsChange = useCallback((columns: ActivistColumnName[]) => {
    dispatch({ type: 'setSelectedColumns', columns })
  }, [])

  const queryOptions = useMemo<QueryActivistOptions>(
    () =>
      buildQueryOptions({
        filters,
        selectedColumns,
        chapterId: user.ChapterID,
        nameSearch: debouncedNameSearch,
      }),
    [filters, selectedColumns, user.ChapterID, debouncedNameSearch],
  )

  const { data, isLoading, isError, error } = useQuery({
    queryKey: [API_PATH.ACTIVISTS_SEARCH, queryOptions],
    queryFn: () => apiClient.searchActivists(queryOptions),
  })

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-1">
        <h1 className="text-2xl font-semibold">Activists</h1>
      </div>

      <ActivistFilters
        filters={filters}
        onFiltersChange={handleFiltersChange}
        isAdmin={isAdmin}
      >
        <ColumnSelector
          visibleColumns={selectedColumns}
          onColumnsChange={handleColumnsChange}
          isChapterColumnShown={filters.searchAcrossChapters}
        />
      </ActivistFilters>

      {isLoading && (
        <div className="flex items-center justify-center py-12 text-muted-foreground">
          Loading activists...
        </div>
      )}

      {isError && (
        <div className="flex items-center justify-center py-12 text-destructive">
          {error instanceof Error
            ? error.message
            : 'Failed to load activists. Please try again.'}
        </div>
      )}

      {data && !isLoading && (
        <>
          <div className="text-sm text-muted-foreground">
            {data.activists.length} activist
            {data.activists.length !== 1 ? 's' : ''} shown
          </div>

          <ActivistTable
            activists={data.activists}
            visibleColumns={selectedColumns}
          />
        </>
      )}
    </div>
  )
}
