'use client'

import { useCallback, useMemo, useState } from 'react'
import { useInfiniteQuery, useQuery } from '@tanstack/react-query'
import { useSearchParams } from 'next/navigation'
import { useQueryState, parseAsInteger } from 'nuqs'
import {
  apiClient,
  API_PATH,
  QueryActivistOptions,
  QueryActivistCountOptions,
  type ActivistColumnName,
  type ActivistJSON,
} from '@/lib/api'
import { activistKeys } from '@/lib/query-keys'
import { useDetectHydrationMismatch } from '@/hooks/use-detect-hydration-mismatch'
import { useAuthedPageContext } from '@/hooks/useAuthedPageContext'
import { InfiniteScrollTrigger } from '@/components/infinite-scroll-trigger'
import { ActivistTable, type ActivistSelection } from './activists-table'
import { SelectionBar } from './selection-bar'
import { BulkAssignDialog } from './bulk-assign-dialog'
import { ActivistFilters } from './filters/activist-filters'
import { ColumnSelector } from './column-selector'
import { SortSelector } from './sort-selector'
import { ActivistSheet } from './activist-sheet'
import { buildQueryOptions } from './filter-api-query'
import type { ActivistsQueryState, SortColumn } from './query-state'
import { DEFAULT_SORT } from './query-state'
import { useActivistQueryState } from './use-activist-query-state'
import { ExportButton } from './export-button'

interface ActivistsPageProps {
  debugInitialServerQueryState?: ActivistsQueryState
  initialReferenceDateIso: string
}

export default function ActivistsPage({
  debugInitialServerQueryState,
  initialReferenceDateIso,
}: ActivistsPageProps) {
  const { user } = useAuthedPageContext()
  const isAdmin = user.Roles.includes('admin')
  // Matches the server's organizer access check on POST /api/activists/assign.
  const canBulkAssign = isAdmin || user.Roles.includes('organizer')
  const searchParams = useSearchParams()
  const isDebug = searchParams.get('debug') === 'true'

  const [selectedActivistId, setSelectedActivistId] = useQueryState(
    'activist',
    parseAsInteger.withOptions({ history: 'push', scroll: false }),
  )

  const {
    filters,
    selectedColumns,
    sort,
    isDirty,
    setFilters,
    setSelectedColumns,
    setSort,
    resetAll,
  } = useActivistQueryState()
  useDetectHydrationMismatch<ActivistsQueryState>({
    label: 'activists query state',
    serverValue: debugInitialServerQueryState,
    clientValue: {
      filters,
      selectedColumns,
      sort,
    },
  })

  const [settledTableState, setSettledTableState] = useState<{
    columns: ActivistColumnName[]
    sort: SortColumn[]
  }>({
    columns: selectedColumns,
    sort,
  })

  const isExplicitSort = sort.length > 0
  const effectiveSort = isExplicitSort ? sort : DEFAULT_SORT
  const initialReferenceDate = useMemo(
    () => new Date(initialReferenceDateIso),
    [initialReferenceDateIso],
  )

  const queryOptions = useMemo<QueryActivistOptions>(
    () =>
      buildQueryOptions({
        filters,
        selectedColumns,
        chapterId: user.ChapterID,
        userId: user.ID,
        referenceDate: initialReferenceDate,
        sort: effectiveSort,
      }),
    [
      filters,
      selectedColumns,
      user.ChapterID,
      user.ID,
      initialReferenceDate,
      effectiveSort,
    ],
  )

  const {
    data,
    isLoading,
    isError,
    error,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    isPlaceholderData,
  } = useInfiniteQuery({
    queryKey: activistKeys.list(queryOptions),
    queryFn: ({ pageParam, signal }) =>
      apiClient.searchActivists(
        {
          ...queryOptions,
          after: pageParam,
        },
        signal,
      ),
    placeholderData: (previousData) => {
      const previousCount =
        previousData?.pages.reduce(
          (total, page) => total + page.activists.length,
          0,
        ) ?? 0

      // Show the previous query's data, if any, while loading data for the new
      // query to avoid having the table disappear completely while new query
      // loads.
      if (previousCount > 0) {
        return previousData
      }

      // If last query returned no results, do not continue showing the message
      // "No activists found matching the current filters." as this could be
      // more easily mistaken for the result of the pending query. Instead, this
      // will show a loading message.
      return undefined
    },
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (lastPage) =>
      lastPage.pagination.next_cursor || undefined,
  })

  const countQueryOptions = useMemo<QueryActivistCountOptions>(
    () => ({ filters: queryOptions.shape.filters }),
    [queryOptions],
  )

  // Not prefetched because it won't cause layout shift and keeps SSR lean.
  const { data: countData, isError: isCountError } = useQuery({
    queryKey: [API_PATH.ACTIVISTS_COUNT, countQueryOptions],
    queryFn: ({ signal }) =>
      apiClient.countActivists(countQueryOptions, signal),
  })

  const activists: ActivistJSON[] = useMemo(
    () => data?.pages.flatMap((page) => page.activists) ?? [],
    [data],
  )

  // Only update the table's columns (and sorting indicators) with those for the
  // new query once the data for that query arrives. This avoids showing
  // the last query's data with the new query's columns.
  if (!isPlaceholderData) {
    const columnsChanged = settledTableState.columns !== selectedColumns
    const sortChanged = settledTableState.sort !== sort
    if (columnsChanged || sortChanged) {
      setSettledTableState({
        columns: selectedColumns,
        sort,
      })
    }
  }

  const tableColumns = isPlaceholderData
    ? settledTableState.columns
    : selectedColumns
  const tableSort = isPlaceholderData ? settledTableState.sort : sort

  const [selectedActivistIds, setSelectedActivistIds] = useState<Set<number>>(
    () => new Set(),
  )
  const [isAssignDialogOpen, setIsAssignDialogOpen] = useState(false)

  const clearSelection = useCallback(() => {
    setSelectedActivistIds((prev) => (prev.size === 0 ? prev : new Set()))
  }, [])

  // A new query returns a different set of rows, so a selection carried over
  // from the old one would be invisible and easy to reassign by accident.
  // Adjusted during render (rather than in an effect) so the dropped selection
  // is never painted alongside the new query's rows.
  const [lastSelectionQueryOptions, setLastSelectionQueryOptions] =
    useState(queryOptions)
  if (lastSelectionQueryOptions !== queryOptions) {
    setLastSelectionQueryOptions(queryOptions)
    clearSelection()
  }

  const toggleActivistSelected = useCallback((id: number) => {
    setSelectedActivistIds((prev) => {
      const next = new Set(prev)
      if (!next.delete(id)) next.add(id)
      return next
    })
  }, [])

  const setManyActivistsSelected = useCallback(
    (ids: number[], selected: boolean) => {
      setSelectedActivistIds((prev) => {
        const next = new Set(prev)
        for (const id of ids) {
          if (selected) next.add(id)
          else next.delete(id)
        }
        return next
      })
    },
    [],
  )

  const selection = useMemo<ActivistSelection | undefined>(
    () =>
      canBulkAssign
        ? {
            selectedIds: selectedActivistIds,
            onToggle: toggleActivistSelected,
            onSetMany: setManyActivistsSelected,
          }
        : undefined,
    [
      canBulkAssign,
      selectedActivistIds,
      toggleActivistSelected,
      setManyActivistsSelected,
    ],
  )

  return (
    <>
      {/* Bounded-height flex chain link (md+) — see frontend-v2/docs/patterns/bounded-height-flex-chain.md */}
      <div
        className={`md:flex-1 md:min-h-0 flex flex-col gap-6 ${
          // Keep the last rows clear of the floating selection bar.
          selectedActivistIds.size > 0 ? 'pb-20' : ''
        }`}
      >
        <div className="flex flex-col gap-1">
          <h1 className="text-2xl font-semibold">Activists</h1>
        </div>

        <ActivistFilters
          filters={filters}
          onFiltersChange={setFilters}
          isAdmin={isAdmin}
          isDirty={isDirty}
          onReset={resetAll}
          exportButton={<ExportButton queryOptions={queryOptions} />}
          isDebug={isDebug}
          debugQueryOptions={queryOptions}
        >
          <ColumnSelector
            visibleColumns={selectedColumns}
            onColumnsChange={setSelectedColumns}
            isChapterColumnShown={filters.searchAcrossChapters}
          />
          <SortSelector
            label="Sort by"
            value={isExplicitSort ? sort[0] : undefined}
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
                {activists.length} of{' '}
                {isCountError ? '?' : (countData?.count ?? '…')} activist
                {(countData?.count ?? 2) !== 1 ? 's' : ''} shown
              </div>
            )}

            <ActivistTable
              activists={activists}
              visibleColumns={tableColumns}
              sort={tableSort}
              onSortChange={setSort}
              onActivistClick={setSelectedActivistId}
              isStale={isPlaceholderData}
              selection={selection}
              footer={
                hasNextPage ? (
                  <InfiniteScrollTrigger
                    onLoadMore={fetchNextPage}
                    isLoading={isFetchingNextPage}
                    canLoadMore={hasNextPage}
                    loadingLabel="Loading more activists…"
                  />
                ) : undefined
              }
            />
          </>
        )}
      </div>

      {canBulkAssign && (
        <>
          <SelectionBar
            count={selectedActivistIds.size}
            onAssign={() => setIsAssignDialogOpen(true)}
            onClear={clearSelection}
          />
          <BulkAssignDialog
            open={isAssignDialogOpen}
            onOpenChange={setIsAssignDialogOpen}
            activistIds={[...selectedActivistIds]}
            onAssigned={clearSelection}
          />
        </>
      )}

      <ActivistSheet
        activistId={selectedActivistId}
        onClose={() => setSelectedActivistId(null)}
      />
    </>
  )
}
