import { useState, useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import { apiClient, API_PATH } from '@/lib/api'
import { ActivistRegistry, type ActivistRecord } from './activist-registry'
import { getActivistStorage } from './activist-storage'
import toast from 'react-hot-toast'

/**
 * Custom hook to access the activist registry with React Query.
 * Loads stored data from IndexedDB and syncs with server in background.
 *
 * Registry manages both in-memory state and IndexedDB storage using write-through pattern.
 *
 * @returns Object containing the registry instance and query state
 */
export function useActivistRegistry(chapterId: number) {
  // Stable registry instance for the lifetime of the component. Held in state
  // (lazy initializer) rather than a ref so it can be read during render.
  const [registry] = useState(() => new ActivistRegistry())
  const [isStorageLoaded, setIsStorageLoaded] = useState(false)
  const [isServerLoaded, setIsServerLoaded] = useState(false)

  // Load stored data on mount (skip if storage is not available)
  useEffect(() => {
    let mounted = true

    const storage = getActivistStorage(chapterId)

    // If IndexedDB is not available (e.g., iOS lockdown mode), skip loading from storage
    if (!storage) {
      console.info(
        '[Registry] IndexedDB not available - running without local caching',
      )
      // One-shot loading flag set in response to an external condition
      // (IndexedDB availability); the extra render on mount is harmless.
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setIsStorageLoaded(true)
      return
    }

    registry
      .loadFromStorage(storage)
      .then(() => {
        if (mounted) setIsStorageLoaded(true)
      })
      .catch(async (err) => {
        console.error('Storage error during loading:', err)

        // Attempt to clear corrupted storage
        try {
          await registry.clearStorage()
        } catch (clearErr) {
          console.error('Failed to clear storage:', clearErr)
        }

        toast.error(
          'Error loading activist cache. Please refresh the page. Contact support if issue persists.',
        )

        if (mounted) setIsStorageLoaded(true)
      })

    return () => {
      mounted = false
    }
  }, [chapterId, registry])

  const query = useQuery({
    queryKey: [API_PATH.ACTIVIST_LIST_BASIC],
    queryFn: async ({ signal }) => {
      // Get last sync time from registry (returns null if storage is unavailable)
      const lastSyncTime = await registry.getLastSyncTime()
      const result = await apiClient.getActivistListBasic(
        lastSyncTime ?? undefined,
        signal,
      )
      return result
    },
    retry: 1,
    staleTime: 5 * 60 * 1000,
    refetchInterval: 10 * 60 * 1000,
  })

  // Handle query errors
  useEffect(() => {
    let mounted = true

    if (query.isError) {
      console.error('[Registry] Server fetch failed:', query.error)
      toast.error(
        'Failed to fetch activist data. Information may be out of date.',
      )
      // One-shot loading flag set in response to an async query error to
      // unblock the UI (with stale data); the extra render is harmless.
      // eslint-disable-next-line react-hooks/set-state-in-effect
      if (mounted) setIsServerLoaded(true)
    }

    return () => {
      mounted = false
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps -- query.error is only read when query.isError is true
  }, [query.isError])

  // Merge server data when it arrives
  useEffect(() => {
    let mounted = true

    if (!query.data) return

    const processServerData = async () => {
      const { activists: newActivists, hidden_ids: hiddenIds } = query.data

      // Remove hidden activists (registry handles both memory and storage)
      if (hiddenIds.length > 0) {
        await registry.removeActivistsByIds(hiddenIds)
      }

      // Filter activists to only those newer than what we have
      const activistsToUpdate: ActivistRecord[] = []

      for (const activist of newActivists) {
        const existingActivist = registry.getActivistById(activist.id)

        if (!existingActivist) {
          activistsToUpdate.push(activist)
          continue
        }

        const hasNewerProfile =
          activist.lastUpdated > existingActivist.lastUpdated
        const hasNewerAttendance =
          activist.lastEventDate > existingActivist.lastEventDate

        if (hasNewerProfile || hasNewerAttendance) {
          activistsToUpdate.push({
            ...existingActivist,
            ...(hasNewerProfile && {
              name: activist.name,
              email: activist.email,
              phone: activist.phone,
              lastUpdated: activist.lastUpdated,
            }),
            ...(hasNewerAttendance && {
              lastEventDate: activist.lastEventDate,
            }),
          })
        }
      }

      // Merge newer activists (registry handles both memory and storage)
      await registry.mergeActivists(activistsToUpdate)

      // Update last sync timestamp
      await registry.setLastSyncTime(new Date().toISOString())

      if (mounted) setIsServerLoaded(true)
    }

    processServerData().catch((err) => {
      console.error('Failed to process server data:', err)
      toast.error(
        'Failed to sync activist data. Information may be out of date.',
      )
      if (mounted) setIsServerLoaded(true) // Mark as loaded to unblock UI
    })

    return () => {
      mounted = false
    }
  }, [query.data, registry])

  return {
    registry,
    isLoading: !isStorageLoaded || !isServerLoaded,
  }
}
