'use client'

import { useCallback, useMemo, useState } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import type { QueryActivistOptions } from '@/lib/api'
import { activistKeys } from '@/lib/query-keys'

/** What a table needs to draw its checkboxes and report clicks on them. */
export interface ActivistSelection {
  selectedIds: ReadonlySet<number>
  /** Flips one activist between selected and not. */
  onToggle: (id: number) => void
  /** Selects or deselects every given activist at once. */
  onSetMany: (ids: number[], selected: boolean) => void
}

export interface ActivistSelectionState {
  /** The selected activists, for callers that need the ids or the count. */
  selectedIds: ReadonlySet<number>
  /** The handlers the table needs to draw and drive its checkboxes. */
  selection: ActivistSelection
  clearSelection: () => void
  /** True once an edit has left the visible rows out of step with the query. */
  areResultsStale: boolean
  markResultsStale: () => void
  /** Refetches the list and clears both the selection and the stale flag. */
  refreshList: () => void
}

/**
 * Tracks which activists are selected for a bulk action, along with whether the
 * rows on screen still match the query that produced them.
 *
 * Both are tied to `queryOptions`: a new query brings its own rows, so anything
 * carried over from the old ones is dropped.
 */
export function useActivistSelection(
  queryOptions: QueryActivistOptions,
): ActivistSelectionState {
  const queryClient = useQueryClient()

  const [selectedActivistIds, setSelectedActivistIds] = useState<Set<number>>(
    () => new Set(),
  )

  // A bulk assign edits the cached rows in place rather than refetching (see
  // bulk-assign-dialog), so rows the assignee filter now excludes stay on
  // screen until the list is refetched. Outlives the selection, which the user
  // may well clear before dealing with the stale rows.
  const [areResultsStale, setAreResultsStale] = useState(false)

  const clearSelection = useCallback(() => {
    setSelectedActivistIds((prev) => (prev.size === 0 ? prev : new Set()))
  }, [])

  const markResultsStale = useCallback(() => setAreResultsStale(true), [])

  // Dropping the selection too, since the rows it points at are the ones most
  // likely to disappear from the refetched list.
  const refreshList = useCallback(() => {
    queryClient.invalidateQueries({ queryKey: activistKeys.lists() })
    // Counted by a separate query, so the total goes stale with the rows.
    queryClient.invalidateQueries({ queryKey: activistKeys.counts() })
    setAreResultsStale(false)
    clearSelection()
  }, [queryClient, clearSelection])

  // A new query returns a different set of rows, so a selection carried over
  // from the old one would be invisible and easy to reassign by accident.
  // Adjusted during render (rather than in an effect) so the dropped selection
  // is never painted alongside the new query's rows.
  const [lastSelectionQueryOptions, setLastSelectionQueryOptions] =
    useState(queryOptions)
  if (lastSelectionQueryOptions !== queryOptions) {
    setLastSelectionQueryOptions(queryOptions)
    clearSelection()
    // A different query fetches its own rows, so nothing carries over as stale.
    setAreResultsStale(false)
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

  const selection = useMemo<ActivistSelection>(
    () => ({
      selectedIds: selectedActivistIds,
      onToggle: toggleActivistSelected,
      onSetMany: setManyActivistsSelected,
    }),
    [selectedActivistIds, toggleActivistSelected, setManyActivistsSelected],
  )

  return {
    selectedIds: selectedActivistIds,
    selection,
    clearSelection,
    areResultsStale,
    markResultsStale,
    refreshList,
  }
}
