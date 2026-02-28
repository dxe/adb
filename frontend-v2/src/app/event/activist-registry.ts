import type { ActivistStorage } from './activist-storage'

export type ActivistRecord = {
  id: number
  name: string
  email: boolean
  phone: boolean
  lastUpdated: number // Unix timestamp in milliseconds
  lastEventDate: number // Unix timestamp in milliseconds, 0 if no events
}

/**
 * In-memory activist registry with write-through storage to IndexedDB.
 *
 * Reads are synchronous (from memory) for fast autocomplete/filtering.
 * Writes are async and automatically persist to IndexedDB when storage is configured.
 *
 * @param storage - Optional IndexedDB storage for persistence.
 *                  When provided, enables automatic write-through caching.
 *                  When omitted, registry operates in memory-only mode.
 */
export class ActivistRegistry {
  private activists: ActivistRecord[]
  private activistsByName: Map<string, ActivistRecord>
  private activistsById: Map<number, ActivistRecord>
  private storage?: ActivistStorage

  constructor(storage?: ActivistStorage) {
    this.activists = []
    this.activistsByName = new Map()
    this.activistsById = new Map()
    this.storage = storage
  }

  /**
   * Sorts activists in-place by lastEventDate descending (most recent first).
   * Activists with no events (lastEventDate === 0) are placed at the end.
   */
  private sortActivists(): void {
    this.activists.sort((a, b) => {
      if (a.lastEventDate === 0 && b.lastEventDate === 0) return 0
      if (a.lastEventDate === 0) return 1 // a has no events, move to end
      if (b.lastEventDate === 0) return -1 // b has no events, move to end
      return b.lastEventDate - a.lastEventDate // Descending (newest first)
    })
  }

  /**
   * Loads activists from IndexedDB storage into memory.
   * Call once after construction, before first use of the registry.
   * @throws Error if storage is not configured
   */
  async loadFromStorage(): Promise<void> {
    if (!this.storage) {
      throw new Error(
        'Cannot load activists from storage: storage not configured.',
      )
    }

    const stored = await this.storage.getAllActivists()
    this.activists = stored
    this.sortActivists()
    this.activistsByName = new Map(stored.map((a) => [a.name, a]))
    this.activistsById = new Map(stored.map((a) => [a.id, a]))
  }

  /**
   * Merges new activists with existing data, replacing duplicates by id.
   * If storage is configured, persists updates to IndexedDB.
   */
  async mergeActivists(newActivists: ActivistRecord[]): Promise<void> {
    if (newActivists.length === 0) {
      return
    }

    const indexById = new Map(this.activists.map((a, i) => [a.id, i]))

    for (const activist of newActivists) {
      const existingIndex = indexById.get(activist.id) ?? -1

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
        indexById.set(activist.id, this.activists.length - 1)
      }

      // Update indexes
      this.activistsByName.set(activist.name, activist)
      this.activistsById.set(activist.id, activist)
    }

    // Re-sort after batch updates to maintain sort order
    this.sortActivists()

    // Write through to storage if configured
    await this.storage?.saveActivists(newActivists)
  }

  /**
   * Removes activists by their IDs from memory and storage.
   * If storage is configured, deletes from IndexedDB.
   */
  async removeActivistsByIds(ids: number[]): Promise<void> {
    const idsToRemove = new Set(ids)

    this.activists = this.activists.filter((activist) => {
      if (idsToRemove.has(activist.id)) {
        this.activistsByName.delete(activist.name)
        this.activistsById.delete(activist.id)
        return false
      }
      return true
    })

    // Write through to storage if configured
    await this.storage?.deleteActivistsByIds(ids)
  }

  /**
   * Gets last sync timestamp from storage.
   */
  async getLastSyncTime(): Promise<string | null> {
    if (!this.storage) return null
    return await this.storage.getLastSyncTime()
  }

  /**
   * Updates last sync timestamp in storage.
   */
  async setLastSyncTime(timestamp: string): Promise<void> {
    await this.storage?.setLastSyncTime(timestamp)
  }

  /**
   * Clears all stored data and reset sync timestamp.
   * Used when storage is corrupted or quota exceeded.
   */
  async clearStorage(): Promise<void> {
    await this.storage?.clearAllActivists()
    await this.storage?.setLastSyncTime(null)
  }

  /**
   * Gets all activists as an array.
   */
  getActivists(): ActivistRecord[] {
    return this.activists
  }

  getActivist(name: string): ActivistRecord | null {
    return this.activistsByName.get(name) ?? null
  }

  getActivistById(id: number): ActivistRecord | null {
    return this.activistsById.get(id) ?? null
  }

  getSuggestions(input: string, maxResults = 10): string[] {
    const trimmedInput = input.trim()
    if (!trimmedInput.length) {
      return []
    }

    // this.activists is pre-sorted by lastEventDate descending
    return this.activists
      .filter(({ name }) => nameFilter(name, input))
      .slice(0, maxResults)
      .map((a) => a.name)
  }

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
