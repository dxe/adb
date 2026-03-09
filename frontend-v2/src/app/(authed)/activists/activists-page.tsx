'use client'

import { useMemo, useEffect, useRef } from 'react'
import { useInfiniteQuery } from '@tanstack/react-query'
import {
  apiClient,
  API_PATH,
  QueryActivistOptions,
  type ActivistJSON,
} from '@/lib/api'
import { useAuthedPageContext } from '@/hooks/useAuthedPageContext'
import { ActivistTable } from './activists-table'
import { ActivistFilters } from './filters/activist-filters'
import { ColumnSelector } from './column-selector'
import { SortSelector } from './sort-selector'
import { buildQueryOptions } from './filter-api-query'
import { useActivistQueryState } from './use-activist-query-state'

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
        // times before the next time React renders this effect.
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

  const {
    filters,
    selectedColumns,
    sort,
    setFilters,
    setSelectedColumns,
    setSort,
  } = useActivistQueryState()

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
    queryFn: ({ pageParam, signal }) =>
      apiClient.searchActivists(
        {
          ...queryOptions,
          after: pageParam,
        },
        signal,
      ),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (lastPage) =>
      lastPage.pagination.next_cursor || undefined,
  })

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
        onFiltersChange={setFilters}
        isAdmin={isAdmin}
      >
        <ColumnSelector
          visibleColumns={selectedColumns}
          onColumnsChange={setSelectedColumns}
          isChapterColumnShown={filters.searchAcrossChapters}
        />
        <SortSelector
          label="Sort by"
          value={sort[0]}
          onChange={(primary) =>
            setSort(
              sort.length > 1 && sort[1].column !== primary.column
                ? [primary, sort[1]]
                : [primary],
            )
          }
          onClear={() => setSort([])}
          canClear={isExplicitSort}
          availableColumns={selectedColumns}
        />
        {isExplicitSort && (
          <SortSelector
            label="Then by"
            inactiveLabel="Then sort by"
            value={sort[1]}
            onChange={(secondary) => setSort([sort[0], secondary])}
            onClear={() => setSort([sort[0]])}
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
            ? error.message.replace(/^invalid query options:\s*/i, '')
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
            onSortChange={setSort}
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
