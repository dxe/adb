import { useState, useEffect, useRef } from 'react'
import { useQuery } from '@tanstack/react-query'
import { apiClient, API_PATH } from '@/lib/api'
import { ActivistRegistry, type ActivistRecord } from './activist-registry'
import { activistStorage } from './activist-storage'

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
  const [isProcessing, setIsProcessing] = useState(false)

  // Load cached data on mount
  useEffect(() => {
    activistStorage
      .getAllActivists()
      .then((cached) => {
        registryRef.current.setActivists(cached)
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
      })
      .finally(() => {
        setIsCacheLoaded(true)
      })
  }, [])

  // Fetch from server with React Query
  const query = useQuery({
    queryKey: [API_PATH.ACTIVIST_LIST_BASIC],
    queryFn: async () => {
      const lastSyncTime = await activistStorage.getLastSyncTime()
      return apiClient.getActivistListBasic(lastSyncTime ?? undefined)
    },
    // Wait for cache to load and prevent concurrent fetches while processing
    enabled: isCacheLoaded && !isProcessing,
    staleTime: 5 * 60 * 1000, // Consider data fresh for 5 minutes
    refetchInterval: 10 * 60 * 1000, // Refetch every 10 minutes in background
  })

  // Merge server data when it arrives
  useEffect(() => {
    if (!query.data || isProcessing) return

    const processServerData = async () => {
      const { activists: newActivists, hidden_ids: hiddenIds } = query.data

      // Remove hidden activists from registry and IndexedDB
      if (hiddenIds.length > 0) {
        registryRef.current.removeActivistsByIds(hiddenIds)
        await activistStorage.deleteActivistsByIds(hiddenIds)
      }

      // Merge new/updated activists into registry and IndexedDB
      if (newActivists.length > 0) {
        registryRef.current.mergeActivists(newActivists)
        // saveActivists uses upsert semantics, so we can just save the new ones
        await activistStorage.saveActivists(newActivists)
      }

      // Update last sync timestamp
      await activistStorage.setLastSyncTime(new Date().toISOString())
    }

    setIsProcessing(true)
    processServerData()
      .catch((err) => {
        console.error('Failed to process server data:', err)
      })
      .finally(() => {
        setIsProcessing(false)
      })
  }, [query.data, isProcessing])

  return {
    registry: registryRef.current,
    isLoading: !isCacheLoaded || query.isLoading,
  }
}
