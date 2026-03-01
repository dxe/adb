'use client'

import { useMemo, useEffect, useCallback, useReducer, useRef } from 'react'
import { useInfiniteQuery } from '@tanstack/react-query'
import { useSearchParams, useRouter } from 'next/navigation'
import {
  apiClient,
  API_PATH,
  QueryActivistOptions,
  ActivistColumnName,
  type ActivistJSON,
} from '@/lib/api'
import { useAuthedPageContext } from '@/hooks/useAuthedPageContext'
import { ActivistTable } from './activists-table'
import { ActivistFilters } from './filters/activist-filters'
import { ColumnSelector } from './column-selector'
import { SortSelector } from './sort-selector'
import {
  normalizeColumnsForFilters,
  columnsForNewFilters,
  DEFAULT_COLUMNS,
} from './column-definitions'
import {
  buildFilterParamEntries,
  buildSortParam,
  parseColumnsFromParams,
  parseFiltersFromParams,
  parseSortFromParams,
} from './filter-url-state'
import {
  type ActivistsQueryState,
  type FilterState,
  type SortColumn,
} from './query-state'
import { buildQueryOptions } from './filter-api-query'

const BASE_PATH = '/activists'

/** Serializes filter/column/sort state to URL search params. */
const buildUrlParams = (
  filters: FilterState,
  visibleColumns: ActivistColumnName[],
  sort: SortColumn[],
): URLSearchParams => {
  const params = new URLSearchParams()

  const filterParams = buildFilterParamEntries(filters)
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

const buildStateFromUrl = (
  getParam: (key: string) => string | undefined,
): ActivistsQueryState => {
  const filters = parseFiltersFromParams(getParam)
  const selectedColumns = normalizeColumnsForFilters(
    parseColumnsFromParams(getParam),
    filters.searchAcrossChapters,
  )
  const sort = parseSortFromParams(getParam, selectedColumns)

  return { filters, selectedColumns, sort }
}

const isExternalNavigation = (
  searchParams: { toString: () => string },
  lastWrittenParams: { current: string },
): boolean => searchParams.toString() !== lastWrittenParams.current

type ActivistsAction =
  | { type: 'setFilters'; filters: FilterState }
  | { type: 'setSelectedColumns'; columns: ActivistColumnName[] }
  | { type: 'setSort'; sort: SortColumn[] }
  | {
      type: 'resetAll'
      filters: FilterState
      selectedColumns: ActivistColumnName[]
      sort: SortColumn[]
    }

const activistsReducer = (
  state: ActivistsQueryState,
  action: ActivistsAction,
): ActivistsQueryState => {
  switch (action.type) {
    case 'setFilters': {
      const { filters } = action
      // Auto-add columns for filters that just got a value set.
      const autoColumns = columnsForNewFilters(state.filters, filters)
      const withAutoColumns =
        autoColumns.length > 0
          ? [...state.selectedColumns, ...autoColumns]
          : state.selectedColumns
      const selectedColumns = normalizeColumnsForFilters(
        withAutoColumns,
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
      return {
        filters: action.filters,
        selectedColumns: action.selectedColumns,
        sort: action.sort,
      }
    }
    default:
      return state
  }
}

function LoadMoreTrigger({
  onLoadMore,
  isLoading,
  canLoadMore,
}: {
  onLoadMore: () => Promise<unknown> | void
  isLoading: boolean
  canLoadMore: boolean
}) {
  const ref = useRef<HTMLDivElement>(null)
  const inFlightRef = useRef(false)

  useEffect(() => {
    const el = ref.current
    if (!el || isLoading || !canLoadMore) return

    const observer = new IntersectionObserver(
      ([entry]) => {
        // isLoading may not have been updated yet. Observer can fire multiple
        // tiems before the next time React renders this effect.
        if (!entry.isIntersecting || inFlightRef.current) return
        inFlightRef.current = true
        void Promise.resolve(onLoadMore()).finally(() => {
          inFlightRef.current = false
        })
      },
      { rootMargin: '200px' },
    )
    observer.observe(el)
    return () => observer.disconnect()
  }, [onLoadMore, isLoading, canLoadMore])

  return (
    <div
      ref={ref}
      className="flex items-center justify-center py-4 text-sm text-muted-foreground"
    >
      <span role="status" aria-live="polite">
        {isLoading ? 'Loading more activists…' : ''}
      </span>
    </div>
  )
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
  const initialState = buildStateFromUrl(getParam)

  const [state, dispatch] = useReducer(activistsReducer, {
    filters: initialState.filters,
    selectedColumns: initialState.selectedColumns,
    sort: initialState.sort,
  })
  const { filters, selectedColumns, sort } = state

  // Track the last URL params we wrote so we can distinguish our own URL
  // updates from external navigation (e.g. clicking a nav preset link).
  const lastWrittenParams = useRef(
    buildUrlParams(filters, selectedColumns, sort).toString(),
  )

  // Reset local UI state when URL changes come from navigation, as if the user
  // was coming to a new "page" even if switching between links in the
  // navigation that come to this same Next.js page. This prevents stale state
  // (such as `visibleFilters` in activist-filters.tsx).
  useEffect(() => {
    if (!isExternalNavigation(searchParams, lastWrittenParams)) return
    const urlState = buildStateFromUrl(getParam)
    dispatch({
      type: 'resetAll',
      filters: urlState.filters,
      selectedColumns: urlState.selectedColumns,
      sort: urlState.sort,
    })
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
        sort,
      }),
    [filters, selectedColumns, user.ChapterID, user.ID, sort],
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

      {!isLoading && !isError && (
        <>
          {activists.length > 0 && (
            <div className="text-sm text-muted-foreground">
              {activists.length} activist
              {activists.length !== 1 ? 's' : ''} shown
            </div>
          )}

          <ActivistTable
            activists={activists}
            visibleColumns={selectedColumns}
            sort={sort}
            onSortChange={handleSortChange}
          />

          {hasNextPage && (
            <LoadMoreTrigger
              onLoadMore={fetchNextPage}
              isLoading={isFetchingNextPage}
              canLoadMore={hasNextPage}
            />
          )}
        </>
      )}
    </div>
  )
}
