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
  // Map to track lastUpdated timestamp for each activist
  const activistTimestampsRef = useRef(new Map<number, number>())
  // Track loading states
  const [isCacheLoaded, setIsCacheLoaded] = useState(false)
  const [isServerLoaded, setIsServerLoaded] = useState(false)

  // Load cached data on mount
  useEffect(() => {
    activistStorage
      .getAllActivists()
      .then((cached) => {
        // Build timestamp map from cached data
        const timestamps = new Map<number, number>()
        for (const activist of cached) {
          timestamps.set(activist.id, activist.lastUpdated)
        }
        activistTimestampsRef.current = timestamps

        // Load into registry (without lastUpdated field)
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
        setIsCacheLoaded(true) // Mark as loaded even on error
      })
  }, [])

  // Fetch from server with React Query (start immediately, don't wait for cache)
  const query = useQuery({
    queryKey: [API_PATH.ACTIVIST_LIST_BASIC],
    queryFn: async () => {
      try {
        const lastSyncTime = await activistStorage.getLastSyncTime()
        const result = await apiClient.getActivistListBasic(
          lastSyncTime ?? undefined,
        )
        return result
      } catch (error) {
        console.error('[Registry] Server fetch failed:', error)
        throw error
      }
    },
    retry: 1, // Only retry once to avoid infinite loops during development
    staleTime: 5 * 60 * 1000, // Consider data fresh for 5 minutes
    refetchInterval: 10 * 60 * 1000, // Refetch every 10 minutes in background
  })

  // Merge server data when it arrives
  useEffect(() => {
    if (!query.data) return

    const processServerData = async () => {
      const { activists: newActivists, hidden_ids: hiddenIds } = query.data

      // Remove hidden activists from registry, IndexedDB, and timestamp map
      if (hiddenIds.length > 0) {
        registryRef.current.removeActivistsByIds(hiddenIds)
        await activistStorage.deleteActivistsByIds(hiddenIds)
        for (const id of hiddenIds) {
          activistTimestampsRef.current.delete(id)
        }
      }

      // Filter activists to only those newer than what we have
      const activistsToUpdate: ActivistRecord[] = []
      const activistsToSave: Array<ActivistRecord & { lastUpdated: number }> =
        []

      for (const activist of newActivists) {
        const existingTimestamp =
          activistTimestampsRef.current.get(activist.id) || 0
        // Server provides Unix timestamp in seconds, convert to milliseconds for comparison
        const activistTimestamp = activist.last_updated * 1000

        // Only update if this data is newer
        if (activistTimestamp > existingTimestamp) {
          // Store without last_updated field for the registry
          const { last_updated, ...activistData } = activist
          activistsToUpdate.push(activistData)
          activistsToSave.push({
            ...activistData,
            lastUpdated: activistTimestamp,
          })
          activistTimestampsRef.current.set(activist.id, activistTimestamp)
        }
      }

      // Merge newer activists into registry and IndexedDB
      if (activistsToUpdate.length > 0) {
        registryRef.current.mergeActivists(activistsToUpdate)
        await activistStorage.saveActivists(activistsToSave)
      }

      // Update last sync timestamp
      await activistStorage.setLastSyncTime(new Date().toISOString())

      setIsServerLoaded(true)
    }

    processServerData().catch((err) => {
      console.error('Failed to process server data:', err)
      setIsServerLoaded(true) // Mark as loaded even on error
    })
  }, [query.data])

  return {
    registry: registryRef.current,
    isLoading: !isCacheLoaded || !isServerLoaded,
  }
}
