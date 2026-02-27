'use client'

import { useMemo, useState, useEffect, useCallback, useReducer, useRef } from 'react'
import { useInfiniteQuery } from '@tanstack/react-query'
import { liteDebounce } from '@tanstack/pacer-lite'
import { useSearchParams, useRouter } from 'next/navigation'
import {
  apiClient,
  API_PATH,
  QueryActivistOptions,
  ActivistColumnName,
  type ActivistJSON,
} from '@/lib/api'
import { Button } from '@/components/ui/button'
import { useAuthedPageContext } from '@/hooks/useAuthedPageContext'
import { ActivistTable } from './activists-table'
import { ActivistFilters } from './filters/activist-filters'
import { ColumnSelector } from './column-selector'
import { SortSelector } from './sort-selector'
import {
  normalizeColumnsForFilters,
  DEFAULT_COLUMNS,
} from './column-definitions'
import {
  FilterState,
  SortColumn,
  buildQueryOptions,
  buildSortParam,
  parseColumnsFromParams,
  parseFiltersFromParams,
  parseSortFromParams,
} from './query-utils'

const BASE_PATH = '/activists'

/** Serializes filter/column/sort state to URL search params. */
const buildUrlParams = (
  filters: FilterState,
  visibleColumns: ActivistColumnName[],
  sort: SortColumn[],
): URLSearchParams => {
  const params = new URLSearchParams()

  const filterParams: [string, string | undefined][] = [
    ['searchAcrossChapters', filters.searchAcrossChapters ? 'true' : undefined],
    ['nameSearch', filters.nameSearch || undefined],
    ['includeHidden', filters.includeHidden ? 'true' : undefined],
    ['lastEvent', filters.lastEvent],
    ['interestDate', filters.interestDate],
    ['firstEvent', filters.firstEvent],
    ['totalEvents', filters.totalEvents],
    ['totalInteractions', filters.totalInteractions],
    ['activistLevel', filters.activistLevel],
    ['source', filters.source],
    ['training', filters.training],
    ['assignedTo', filters.assignedTo],
    ['followups', filters.followups],
    ['prospect', filters.prospect],
  ]
  for (const [key, value] of filterParams) {
    if (value) params.set(key, value)
  }

  const defaultColumns = normalizeColumnsForFilters(
    DEFAULT_COLUMNS,
    filters.searchAcrossChapters,
  )

  // Only include columns in URL if they differ from defaults
  const columnsMatch =
    visibleColumns.length === defaultColumns.length &&
    visibleColumns.every((col, idx) => col === defaultColumns[idx])

  if (!columnsMatch) {
    const columnsToStore = visibleColumns.filter(
      (col) => col !== 'chapter_name' && col !== 'name',
    )
    if (columnsToStore.length > 0) {
      params.set('columns', columnsToStore.join(','))
    }
  }

  const sortParam = buildSortParam(sort)
  if (sortParam) {
    params.set('sort', sortParam)
  }

  return params
}

type ActivistsState = {
  filters: FilterState
  selectedColumns: ActivistColumnName[]
  sort: SortColumn[]
}

type ActivistsAction =
  | { type: 'setFilters'; filters: FilterState }
  | { type: 'setSelectedColumns'; columns: ActivistColumnName[] }
  | { type: 'setSort'; sort: SortColumn[] }
  | { type: 'resetAll'; filters: FilterState; selectedColumns: ActivistColumnName[]; sort: SortColumn[] }

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
      return { ...state, filters, selectedColumns }
    }
    case 'setSelectedColumns': {
      const selectedColumns = normalizeColumnsForFilters(
        action.columns,
        state.filters.searchAcrossChapters,
      )

      // Reset sorting when sorted column is unselected
      const sort = state.sort.every((s) => selectedColumns.includes(s.column))
        ? state.sort
        : []

      return { ...state, selectedColumns, sort }
    }
    case 'setSort': {
      return { ...state, sort: action.sort }
    }
    case 'resetAll': {
      return { filters: action.filters, selectedColumns: action.selectedColumns, sort: action.sort }
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
  const initialColumns = normalizeColumnsForFilters(
    parseColumnsFromParams(getParam),
    initialFilters.searchAcrossChapters,
  )

  const initialSort = parseSortFromParams(getParam, initialColumns)

  const [state, dispatch] = useReducer(activistsReducer, {
    filters: initialFilters,
    selectedColumns: initialColumns,
    sort: initialSort,
  })
  const { filters, selectedColumns, sort } = state

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

  // Track the last URL params we wrote so we can distinguish our own URL
  // updates from external navigation (e.g. clicking a nav preset link).
  const lastWrittenParams = useRef(
    buildUrlParams(filters, selectedColumns, sort).toString(),
  )

  // Sync state from URL on external navigation (e.g. nav preset links).
  // When the user clicks a Link to the same route with different params,
  // React doesn't remount the component, so the reducer keeps stale state.
  useEffect(() => {
    const currentParams = searchParams.toString()
    if (currentParams === lastWrittenParams.current) return
    const urlFilters = parseFiltersFromParams(getParam)
    const urlColumns = normalizeColumnsForFilters(
      parseColumnsFromParams(getParam),
      urlFilters.searchAcrossChapters,
    )
    const urlSort = parseSortFromParams(getParam, urlColumns)
    dispatch({ type: 'resetAll', filters: urlFilters, selectedColumns: urlColumns, sort: urlSort })
    setDebouncedNameSearch(urlFilters.nameSearch)
  }, [searchParams, getParam])

  // Update URL when filters, columns, or sort change
  useEffect(() => {
    const params = buildUrlParams(filters, selectedColumns, sort).toString()
    lastWrittenParams.current = params
    const newUrl = params ? `?${params}` : BASE_PATH
    router.replace(newUrl, { scroll: false })
  }, [filters, selectedColumns, sort, router])

  const handleFiltersChange = useCallback((newFilters: FilterState) => {
    dispatch({ type: 'setFilters', filters: newFilters })
  }, [])

  const handleColumnsChange = useCallback((columns: ActivistColumnName[]) => {
    dispatch({ type: 'setSelectedColumns', columns })
  }, [])

  const handleSortChange = useCallback((newSort: SortColumn[]) => {
    dispatch({ type: 'setSort', sort: newSort })
  }, [])

  const isExplicitSort = sort.length > 0

  const queryOptions = useMemo<QueryActivistOptions>(
    () =>
      buildQueryOptions({
        filters,
        selectedColumns,
        chapterId: user.ChapterID,
        userId: user.ID,
        nameSearch: debouncedNameSearch,
        sort,
      }),
    [filters, selectedColumns, user.ChapterID, user.ID, debouncedNameSearch, sort],
  )

  const {
    data,
    isLoading,
    isError,
    error,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
  } = useInfiniteQuery({
    queryKey: [API_PATH.ACTIVISTS_SEARCH, queryOptions],
    queryFn: ({ pageParam }) =>
      apiClient.searchActivists({
        ...queryOptions,
        after: pageParam,
      }),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (lastPage) =>
      lastPage.pagination.next_cursor || undefined,
  })

  // Flatten pages of activists into one array
  const activists: ActivistJSON[] = useMemo(
    () => data?.pages.flatMap((page) => page.activists) ?? [],
    [data],
  )

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
        <SortSelector
          label="Sort by"
          value={sort[0]}
          onChange={(primary) =>
            handleSortChange(
              sort.length > 1 && sort[1].column !== primary.column
                ? [primary, sort[1]]
                : [primary],
            )
          }
          onClear={() => handleSortChange([])}
          canClear={isExplicitSort}
          availableColumns={selectedColumns}
        />
        {isExplicitSort && (
          <SortSelector
            label="Then by"
            inactiveLabel="Then sort by"
            value={sort[1]}
            onChange={(secondary) => handleSortChange([sort[0], secondary])}
            onClear={() => handleSortChange([sort[0]])}
            availableColumns={selectedColumns.filter(
              (col) => col !== sort[0].column,
            )}
          />
        )}
      </ActivistFilters>

      {isLoading && (
        <div className="flex items-center justify-center py-12 text-muted-foreground">
          Loading activists...
        </div>
      )}

      {isError && (
        <div className="flex items-center justify-center py-12 text-destructive">
          {error instanceof Error
            ? error.message.replace(/^invalid query options:\s*/i, '') // Remove message prefix / boilerplate
            : 'Failed to load activists. Please try again.'}
        </div>
      )}

      {activists.length > 0 && !isLoading && (
        <>
          <div className="text-sm text-muted-foreground">
            {activists.length} activist
            {activists.length !== 1 ? 's' : ''} shown
          </div>

          <ActivistTable
            activists={activists}
            visibleColumns={selectedColumns}
            sort={sort}
            onSortChange={handleSortChange}
          />

          {hasNextPage && (
            <Button
              variant="outline"
              className="self-center"
              onClick={() => fetchNextPage()}
              disabled={isFetchingNextPage}
            >
              {isFetchingNextPage ? 'Loading...' : 'Load more'}
            </Button>
          )}
        </>
      )}
    </div>
  )
}
