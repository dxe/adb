import { useState, useEffect, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { apiClient } from '@/lib/api'
import { ActivistRegistry, type ActivistRecord } from './activist-registry'
import { activistStorage } from './activist-storage'

async function loadCachedActivists(): Promise<ActivistRecord[]> {
  try {
    return await activistStorage.getAllActivists()
  } catch (error) {
    console.error('Failed to load cached activists:', error)
    return []
  }
}

async function fetchActivists() {
  const lastSyncTime = await activistStorage.getLastSyncTime()
  return apiClient.getActivistListBasic(lastSyncTime ?? undefined)
}

/**
 * Custom hook to access the activist registry with React Query.
 * Loads cached data from IndexedDB and syncs with server in background.
 *
 * @returns Object containing the registry instance and query state
 */
export function useActivistRegistry() {
  // Local state for activist data
  const [activists, setActivists] = useState<ActivistRecord[]>([])
  const [isCacheLoaded, setIsCacheLoaded] = useState(false)

  // Load cached data on mount
  useEffect(() => {
    loadCachedActivists().then((cached) => {
      setActivists(cached)
      setIsCacheLoaded(true)
    })
  }, [])

  // Fetch from server with React Query
  const query = useQuery({
    queryKey: ['activists'],
    queryFn: fetchActivists,
    enabled: isCacheLoaded, // Wait for cache to load first
    staleTime: 5 * 60 * 1000, // Consider data fresh for 5 minutes
    refetchInterval: 10 * 60 * 1000, // Refetch every 10 minutes in background
  })

  // Merge server data when it arrives
  useEffect(() => {
    if (!query.data) return

    const processServerData = async () => {
      const { activists: newActivists, hidden_ids: hiddenIds } = query.data

      // Create temporary registry to perform merge operations
      const tempRegistry = new ActivistRegistry(activists)

      // Delete hidden activists
      if (hiddenIds.length > 0) {
        await activistStorage.deleteActivistsByIds(hiddenIds)
        tempRegistry.removeActivistsByIds(hiddenIds)
      }

      // Merge new/updated activists
      if (newActivists.length > 0) {
        tempRegistry.mergeActivists(newActivists)
      }

      // Get merged data and update state
      const mergedActivists = tempRegistry.getActivists()
      setActivists(mergedActivists)

      // Save to IndexedDB
      await activistStorage.saveActivists(mergedActivists)

      // Update last sync timestamp
      await activistStorage.setLastSyncTime(new Date().toISOString())
    }

    processServerData().catch((error) => {
      console.error('Failed to process server data:', error)
    })
  }, [query.data, activists])

  // Create registry from current activist data (memoized)
  const registry = useMemo(() => new ActivistRegistry(activists), [activists])

  return {
    registry,
    isLoading: !isCacheLoaded || query.isLoading,
  }
}
