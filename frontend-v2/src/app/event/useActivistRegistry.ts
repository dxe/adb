import { useState, useEffect, useRef } from 'react'
import { useQuery } from '@tanstack/react-query'
import { apiClient, API_PATH } from '@/lib/api'
import { ActivistRegistry, type ActivistRecord } from './activist-registry'
import { activistStorage } from './activist-storage'
import toast from 'react-hot-toast'

/**
 * Custom hook to access the activist registry with React Query.
 * Loads stored data from IndexedDB and syncs with server in background.
 *
 * Registry manages both in-memory state and IndexedDB storage using write-through pattern.
 *
 * @returns Object containing the registry instance and query state
 */
export function useActivistRegistry() {
  // Single registry instance with write-through storage to IndexedDB (if available)
  const registryRef = useRef(new ActivistRegistry(activistStorage))
  const [isStorageLoaded, setIsStorageLoaded] = useState(false)
  const [isServerLoaded, setIsServerLoaded] = useState(false)

  // Load stored data on mount (skip if storage is not available)
  useEffect(() => {
    let mounted = true

    // If IndexedDB is not available (e.g., iOS lockdown mode), skip loading from storage
    if (!activistStorage) {
      console.info(
        '[Registry] IndexedDB not available - running without local caching',
      )
      setIsStorageLoaded(true)
      return
    }

    registryRef.current
      .loadFromStorage()
      .then(() => {
        if (mounted) setIsStorageLoaded(true)
      })
      .catch(async (err) => {
        console.error('Storage error during loading:', err)

        // Attempt to clear corrupted storage
        try {
          await registryRef.current.clearStorage()
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
  }, [])

  const query = useQuery({
    queryKey: [API_PATH.ACTIVIST_LIST_BASIC],
    queryFn: async () => {
      // Get last sync time from registry (returns null if storage is unavailable)
      const lastSyncTime = await registryRef.current.getLastSyncTime()
      const result = await apiClient.getActivistListBasic(
        lastSyncTime ?? undefined,
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
      if (mounted) setIsServerLoaded(true) // Mark as loaded to unblock UI (with stale data)
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
        await registryRef.current.removeActivistsByIds(hiddenIds)
      }

      // Filter activists to only those newer than what we have
      const activistsToUpdate: ActivistRecord[] = []

      for (const activist of newActivists) {
        const incomingTimestamp = activist.last_updated * 1000
        const existingActivist = registryRef.current.getActivistById(
          activist.id,
        )
        const existingTimestamp = existingActivist?.lastUpdated || 0
        const hasNewAttendanceData =
          activist.last_event_date * 1000 >
          (existingActivist?.lastEventDate || 0)

        // Only update if profile data OR event data is newer (handles out-of-order responses)
        if (incomingTimestamp > existingTimestamp || hasNewAttendanceData) {
          const { last_event_date, ...activistData } = activist
          activistsToUpdate.push({
            ...activistData,
            lastUpdated: incomingTimestamp,
            lastEventDate: last_event_date * 1000,
          })
        }
      }

      // Merge newer activists (registry handles both memory and storage)
      if (activistsToUpdate.length > 0) {
        await registryRef.current.mergeActivists(activistsToUpdate)
      }

      // Update last sync timestamp
      await registryRef.current.setLastSyncTime(new Date().toISOString())

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
  }, [query.data])

  return {
    registry: registryRef.current,
    isLoading: !isStorageLoaded || !isServerLoaded,
  }
}
