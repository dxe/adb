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
  // Single registry instance with write-through storage to IndexedDB
  const registryRef = useRef(new ActivistRegistry(activistStorage))
  const [isStorageLoaded, setIsStorageLoaded] = useState(false)
  const [isServerLoaded, setIsServerLoaded] = useState(false)

  // Load stored data on mount
  useEffect(() => {
    let mounted = true

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
          'Error loading activists. Please refresh the page. Contact support if issue persists.',
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
      const lastSyncTime = await activistStorage.getLastSyncTime()
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
    if (query.isError) {
      console.error('[Registry] Server fetch failed:', query.error)
      toast.error(
        'Failed to fetch activist data. Information may be out of date.',
      )
      setIsServerLoaded(true) // Mark as loaded to unblock UI (with stale data)
    }
  }, [query.isError, query.error])

  // Merge server data when it arrives
  useEffect(() => {
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
        // Server provides Unix timestamp in seconds, convert to milliseconds
        const incomingTimestamp = activist.last_updated * 1000
        const existingActivist = registryRef.current.getActivistById(
          activist.id,
        )
        const existingTimestamp = existingActivist?.lastUpdated || 0

        // Only update if this data is newer (handles out-of-order responses)
        if (incomingTimestamp > existingTimestamp) {
          const { last_updated, ...activistData } = activist
          activistsToUpdate.push({
            ...activistData,
            lastUpdated: incomingTimestamp,
          })
        }
      }

      // Merge newer activists (registry handles both memory and storage)
      if (activistsToUpdate.length > 0) {
        await registryRef.current.mergeActivists(activistsToUpdate)
      }

      // Update last sync timestamp
      await registryRef.current.setLastSyncTime(new Date().toISOString())

      setIsServerLoaded(true)
    }

    processServerData().catch((err) => {
      console.error('Failed to process server data:', err)
      toast.error(
        'Failed to sync activist data. Information may be out of date.',
      )
      setIsServerLoaded(true) // Mark as loaded to unblock UI
    })
  }, [query.data])

  return {
    registry: registryRef.current,
    isLoading: !isStorageLoaded || !isServerLoaded,
  }
}
