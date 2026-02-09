'use client'

import { useMemo, useState, useEffect } from 'react'
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
import { ActivistFilters, FilterState } from './activist-filters'
import { ColumnSelector } from './column-selector'
import { getDefaultColumns, sortColumnsByDefinitionOrder } from './column-definitions'

export default function ActivistsPage() {
  const { user } = useAuthedPageContext()
  const isAdmin = user.role === 'admin'
  const searchParams = useSearchParams()
  const router = useRouter()

  // Parse initial state from URL
  const getInitialFilters = (): FilterState => {
    return {
      showAllChapters: searchParams.get('showAllChapters') === 'true',
      nameSearch: searchParams.get('nameSearch') || '',
      lastEventGte: searchParams.get('lastEventGte') || undefined,
      lastEventLt: searchParams.get('lastEventLt') || undefined,
    }
  }

  const getInitialColumns = (): ActivistColumnName[] => {
    const columnsParam = searchParams.get('columns')
    if (columnsParam) {
      const parsed = columnsParam.split(',') as ActivistColumnName[]
      return sortColumnsByDefinitionOrder(parsed)
    }
    return getDefaultColumns(false)
  }

  // Filter state
  const [filters, setFilters] = useState<FilterState>(getInitialFilters)

  // Column state
  const [visibleColumns, setVisibleColumns] =
    useState<ActivistColumnName[]>(getInitialColumns)

  // Debounced name search (for API queries)
  const [debouncedNameSearch, setDebouncedNameSearch] = useState(
    filters.nameSearch,
  )

  // Create debounced setter with @tanstack/pacer-lite
  const debouncedSetNameSearch = useMemo(
    () =>
      liteDebounce((value: string) => setDebouncedNameSearch(value), {
        wait: 300,
      }),
    [],
  )

  // Update debounced search when filters change
  useEffect(() => {
    debouncedSetNameSearch(filters.nameSearch)
  }, [filters.nameSearch, debouncedSetNameSearch])

  // Update URL when filters or columns change
  useEffect(() => {
    const params = new URLSearchParams()

    // Add filter params
    if (filters.showAllChapters) {
      params.set('showAllChapters', 'true')
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

    // Add columns param (exclude chapter_name and name as they're auto-added)
    const columnsToStore = visibleColumns.filter(
      (col) => col !== 'chapter_name' && col !== 'name'
    )
    if (columnsToStore.length > 0) {
      params.set('columns', columnsToStore.join(','))
    }

    // Update URL without causing navigation
    const newUrl = params.toString() ? `?${params.toString()}` : '/activists'
    router.replace(newUrl, { scroll: false })
  }, [filters, visibleColumns, router])

  // Build query options (using debounced search)
  const queryOptions = useMemo<QueryActivistOptions>(() => {
    // Determine columns to request from API
    const columnsToRequest = [...visibleColumns]

    // Add chapter_name if showing all chapters and not already included
    if (filters.showAllChapters && !columnsToRequest.includes('chapter_name')) {
      columnsToRequest.unshift('chapter_name')
    }

    // Always include ID for row keys
    if (!columnsToRequest.includes('id')) {
      columnsToRequest.unshift('id')
    }

    return {
      columns: columnsToRequest,
      filters: {
        chapter_id: filters.showAllChapters ? 0 : user.ChapterID,
        name: debouncedNameSearch
          ? { name_contains: debouncedNameSearch }
          : undefined,
        last_event:
          filters.lastEventGte || filters.lastEventLt
            ? {
                last_event_gte: filters.lastEventGte,
                last_event_lt: filters.lastEventLt,
              }
            : undefined,
      },
    }
  }, [
    filters.showAllChapters,
    filters.lastEventGte,
    filters.lastEventLt,
    debouncedNameSearch,
    visibleColumns,
    user.ChapterID,
  ])

  // Fetch activists
  const { data, isLoading, isError, error } = useQuery({
    queryKey: [API_PATH.ACTIVISTS_SEARCH, queryOptions],
    queryFn: () => apiClient.searchActivists(queryOptions),
  })

  // Update visible columns when showAllChapters changes
  const handleFiltersChange = (newFilters: FilterState) => {
    setFilters(newFilters)

    // Only update chapter_name column visibility, preserve other selections
    if (newFilters.showAllChapters !== filters.showAllChapters) {
      setVisibleColumns((currentCols) => {
        const hasChapterName = currentCols.includes('chapter_name')

        if (newFilters.showAllChapters && !hasChapterName) {
          // Add chapter_name at the beginning
          return ['chapter_name', ...currentCols]
        } else if (!newFilters.showAllChapters && hasChapterName) {
          // Remove chapter_name
          return currentCols.filter((col) => col !== 'chapter_name')
        }

        return currentCols
      })
    }
  }

  // Get display columns (for table rendering)
  const displayColumns = useMemo<ActivistColumnName[]>(() => {
    const cols = [...visibleColumns]

    // Add chapter_name if showing all chapters and not already included
    if (filters.showAllChapters && !cols.includes('chapter_name')) {
      cols.unshift('chapter_name')
    }

    return cols
  }, [visibleColumns, filters.showAllChapters])

  return (
    <div className="flex flex-col gap-6">
      {/* Header */}
      <div className="flex flex-col gap-1">
        <h1 className="text-2xl font-semibold">Activists</h1>
        <p className="text-sm text-muted-foreground">
          Search and manage activists in your chapter
        </p>
      </div>

      {/* Filters */}
      <ActivistFilters
        filters={filters}
        onFiltersChange={handleFiltersChange}
        isAdmin={isAdmin}
      >
        <ColumnSelector
          visibleColumns={visibleColumns}
          onColumnsChange={setVisibleColumns}
          showAllChapters={filters.showAllChapters}
        />
      </ActivistFilters>

      {/* Loading state */}
      {isLoading && (
        <div className="flex items-center justify-center py-12 text-muted-foreground">
          Loading activists...
        </div>
      )}

      {/* Error state */}
      {isError && (
        <div className="flex items-center justify-center py-12 text-destructive">
          {error instanceof Error
            ? error.message
            : 'Failed to load activists. Please try again.'}
        </div>
      )}

      {/* Results count */}
      {data && !isLoading && (
        <div className="text-sm text-muted-foreground">
          {data.activists.length} activist
          {data.activists.length !== 1 ? 's' : ''} shown
        </div>
      )}

      {/* Table */}
      {data && !isLoading && (
        <ActivistTable
          activists={data.activists}
          visibleColumns={displayColumns}
        />
      )}
    </div>
  )
}
