import { useState, useEffect, useRef } from 'react'
import { useQuery } from '@tanstack/react-query'
import { apiClient, API_PATH } from '@/lib/api'
import { ActivistRegistry, type ActivistRecord } from './activist-registry'
import { activistStorage } from './activist-storage'
import toast from 'react-hot-toast'

/**
 * Custom hook to access the activist registry with React Query.
 * Loads cached data from IndexedDB and syncs with server in background.
 *
 * @returns Object containing the registry instance and query state
 */
export function useActivistRegistry() {
  // Single registry instance updated in place (not recreated on every change)
  const registryRef = useRef(new ActivistRegistry())
  const [isCacheLoaded, setIsCacheLoaded] = useState(false)
  const [isServerLoaded, setIsServerLoaded] = useState(false)

  // Load cached data on mount
  useEffect(() => {
    activistStorage
      .getAllActivists()
      .then((cached) => {
        registryRef.current.setActivists(cached)
        setIsCacheLoaded(true)
      })
      .catch(async (err) => {
        console.error('Error loading cached activists, clearing storage:', err)
        // Clear corrupted cache and force full refresh from server
        try {
          await activistStorage.clearAllActivists()
          await activistStorage.setLastSyncTime('')
        } catch (clearErr) {
          console.error('Error clearing corrupted storage:', clearErr)
          // If clearing fails, IndexedDB might be permanently broken
          // (quota exceeded, disk full, browser issues). We silently degrade
          // to no-cache mode - the app still works via server fetches, just slower.
          // This is preferable to blocking the user completely.
        }
        setIsCacheLoaded(true)
      })
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
      // Show error to user if we have no cached data
      if (registryRef.current.size() === 0) {
        toast.error('Failed to fetch activist data. Please refresh the page.')
      } else {
        toast.error(
          'Failed to fetch latest activist data. Activist data may be out of date.',
        )
      }
      setIsServerLoaded(true) // Mark as loaded to unblock UI (with stale data)
    }
  }, [query.isError, query.error])

  // Merge server data when it arrives
  useEffect(() => {
    if (!query.data) return

    const processServerData = async () => {
      const { activists: newActivists, hidden_ids: hiddenIds } = query.data

      // Remove hidden activists from registry and IndexedDB
      if (hiddenIds.length > 0) {
        registryRef.current.removeActivistsByIds(hiddenIds)
        await activistStorage.deleteActivistsByIds(hiddenIds)
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

      // Merge newer activists into registry and IndexedDB
      if (activistsToUpdate.length > 0) {
        registryRef.current.mergeActivists(activistsToUpdate)
        await activistStorage.saveActivists(activistsToUpdate)
      }

      // Update last sync timestamp
      await activistStorage.setLastSyncTime(new Date().toISOString())

      setIsServerLoaded(true)
    }

    processServerData().catch((err) => {
      console.error('Failed to process server data:', err)
      // Show error to user if we have no cached data
      if (registryRef.current.size() === 0) {
        toast.error('Failed to sync activist data. Please refresh the page.')
      } else {
        toast.error(
          'Failed to sync latest activist data. Activist data may be out of date.',
        )
      }
      setIsServerLoaded(true) // Mark as loaded to unblock UI
    })
  }, [query.data])

  return {
    registry: registryRef.current,
    isLoading: !isCacheLoaded || !isServerLoaded,
  }
}
