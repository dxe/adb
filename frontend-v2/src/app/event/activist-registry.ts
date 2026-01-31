import { apiClient } from '@/lib/api'
import {
  deleteActivistsByIds,
  getAllActivists,
  getLastSyncTime,
  saveActivists,
  setLastSyncTime,
} from './activist-storage'

type ActivistRecord = {
  id: number
  name: string
  email?: string
  phone?: string
}

/**
 * Encapsulates basic activist data and provides lookup methods.
 * Supports IndexedDB caching and incremental syncing.
 */
export class ActivistRegistry {
  private activists: ActivistRecord[]
  private activistsByName: Map<string, ActivistRecord>
  private isInitialized = false

  constructor(activists: ActivistRecord[] = []) {
    this.activists = activists
    this.activistsByName = new Map(activists.map((a) => [a.name, a]))
  }

  /**
   * Initialize the registry by loading cached data from IndexedDB,
   * then fetch any new/updated activists from the server.
   */
  static async create(): Promise<ActivistRegistry> {
    const registry = new ActivistRegistry()
    await registry.initialize()
    return registry
  }

  /**
   * Load activists from IndexedDB cache and sync with server.
   */
  private async initialize(): Promise<void> {
    try {
      // Load cached activists from IndexedDB
      const cachedActivists = await getAllActivists()
      this.setActivists(cachedActivists)

      // Fetch and merge any new/updated activists
      await this.sync()

      this.isInitialized = true
    } catch (error) {
      console.error('Failed to initialize activist registry:', error)
      // Fall back to empty registry on error
      this.setActivists([])
      this.isInitialized = true
    }
  }

  /**
   * Fetch activists modified since last sync and merge with cached data.
   */
  async sync(): Promise<void> {
    try {
      const lastSyncTime = await getLastSyncTime()
      const currentTime = new Date().toISOString()

      // Fetch new/updated activists from server
      const response = await apiClient.getActivistListBasic(
        lastSyncTime ?? undefined,
      )
      const newActivists = response.activists
      const hiddenIds = response.hidden_ids

      // Delete hidden activists from cache and memory
      if (hiddenIds.length > 0) {
        await deleteActivistsByIds(hiddenIds)
        this.removeActivistsByIds(hiddenIds)
      }

      if (newActivists.length > 0) {
        // Merge with existing data (upsert semantics by id)
        this.mergeActivists(newActivists)

        // Save merged data to IndexedDB
        await saveActivists(this.activists)
      }

      // Update last sync timestamp
      await setLastSyncTime(currentTime)
    } catch (error) {
      console.error('Failed to sync activists:', error)
      throw error
    }
  }

  /**
   * Replace all activists with new data and rebuild index.
   */
  private setActivists(activists: ActivistRecord[]): void {
    this.activists = activists
    this.activistsByName = new Map(activists.map((a) => [a.name, a]))
  }

  /**
   * Merge new activists with existing data, replacing duplicates by id.
   */
  private mergeActivists(newActivists: ActivistRecord[]): void {
    // Update existing activists and add new ones
    for (const activist of newActivists) {
      const existingIndex = this.activists.findIndex(
        (a) => a.id === activist.id,
      )

      if (existingIndex >= 0) {
        // Update existing activist (handles renames properly)
        const oldActivist = this.activists[existingIndex]
        this.activists[existingIndex] = activist

        // Remove old name from index if name changed
        if (oldActivist.name !== activist.name) {
          this.activistsByName.delete(oldActivist.name)
        }
      } else {
        // Add new activist
        this.activists.push(activist)
      }

      // Update name index
      this.activistsByName.set(activist.name, activist)
    }
  }

  /**
   * Remove activists by their IDs from memory.
   */
  private removeActivistsByIds(ids: number[]): void {
    const idsToRemove = new Set(ids)

    // Remove from activists array
    this.activists = this.activists.filter((activist) => {
      if (idsToRemove.has(activist.id)) {
        // Remove from name index
        this.activistsByName.delete(activist.name)
        return false
      }
      return true
    })
  }

  getActivist(name: string): ActivistRecord | null {
    return this.activistsByName.get(name) ?? null
  }

  getSuggestions(input: string, maxResults = 10): string[] {
    const trimmedInput = input.trim()
    if (!trimmedInput.length) {
      return []
    }

    return this.activists
      .filter(({ name }) => nameFilter(name, input))
      .slice(0, maxResults)
      .map((a) => a.name)
  }

  /**
   * Check if the registry has been initialized.
   */
  isReady(): boolean {
    return this.isInitialized
  }

  /**
   * Get the total number of activists in the registry.
   */
  size(): number {
    return this.activists.length
  }
}

/**
 * Filters text based on a flexible name matching pattern.
 * Treats whitespace as a wildcard allowing any characters in between,
 * enabling partial and out-of-order matching (e.g., "john doe" matches "John Q. Doe").
 * Matching is case-insensitive.
 *
 * @param text - The text to search within
 * @param input - The search pattern
 * @returns true if the pattern matches the text
 */
function nameFilter(text: string, input: string): boolean {
  const pattern = input
    .trim()
    .replace(/[.*+?^${}()|[\]\\]/g, '\\$&') // escape special regex chars
    .replace(/ +/g, '.*') // whitespace matches anything
  return new RegExp(pattern, 'i').test(text)
}
