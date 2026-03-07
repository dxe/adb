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
 * When storage is not configured, registry operates in memory-only mode.
 */
export class ActivistRegistry {
  private activists: ActivistRecord[]
  private activistsByName: Map<string, ActivistRecord>
  private activistsById: Map<number, ActivistRecord>
  private storage?: ActivistStorage

  constructor() {
    this.activists = []
    this.activistsByName = new Map()
    this.activistsById = new Map()
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
   */
  async loadFromStorage(storage: ActivistStorage): Promise<void> {
    this.storage = storage

    this.activists = await this.storage.getAllActivists()
    this.sortActivists()
    this.activistsByName = new Map(this.activists.map((a) => [a.name, a]))
    this.activistsById = new Map(this.activists.map((a) => [a.id, a]))
  }

  /**
   * Merges new activists with existing data.
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
        this.activists[existingIndex] = activist
      } else {
        this.activists.push(activist)
        indexById.set(activist.id, this.activists.length - 1)
      }

      this.activistsById.set(activist.id, activist)
    }

    this.sortActivists()
    this.activistsByName = new Map(this.activists.map((a) => [a.name, a]))

    // Write through to storage if configured
    await this.storage?.saveActivists(newActivists)
  }

  /**
   * Removes activists by their IDs from memory and storage.
   * If storage is configured, deletes from IndexedDB.
   */
  async removeActivistsByIds(ids: number[]): Promise<void> {
    if (ids.length === 0) {
      return
    }

    const idsToRemove = new Set(ids)

    const remainingActivists: ActivistRecord[] = []
    for (const activist of this.activists) {
      if (idsToRemove.has(activist.id)) {
        this.activistsByName.delete(activist.name)
        this.activistsById.delete(activist.id)
        continue
      }
      remainingActivists.push(activist)
    }
    this.activists = remainingActivists

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

    // Build regex once for flexible name matching pattern:
    // - Treats whitespace as a wildcard allowing any characters in between
    // - Enables partial and out-of-order matching (e.g., "john doe" matches "John Q. Doe")
    // - Matching is case-insensitive
    const pattern = trimmedInput
      .replace(/[.*+?^${}()|[\]\\]/g, '\\$&') // escape special regex chars
      .replace(/ +/g, '.*') // whitespace matches anything
    const regex = new RegExp(pattern, 'i')

    // this.activists is pre-sorted by lastEventDate descending
    return this.activists
      .filter(({ name }) => regex.test(name))
      .slice(0, maxResults)
      .map((a) => a.name)
  }

  size(): number {
    return this.activists.length
  }
}
