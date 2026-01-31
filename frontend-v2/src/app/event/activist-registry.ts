export type ActivistRecord = {
  id: number
  name: string
  email?: string
  phone?: string
}

/**
 * Pure data structure for activist lookups and filtering.
 * No fetching logic - that's handled by React Query.
 */
export class ActivistRegistry {
  private activists: ActivistRecord[]
  private activistsByName: Map<string, ActivistRecord>

  constructor(activists: ActivistRecord[] = []) {
    this.activists = activists
    this.activistsByName = new Map(activists.map((a) => [a.name, a]))
  }

  /**
   * Replace all activists with new data and rebuild index.
   */
  setActivists(activists: ActivistRecord[]): void {
    this.activists = activists
    this.activistsByName = new Map(activists.map((a) => [a.name, a]))
  }

  /**
   * Merge new activists with existing data, replacing duplicates by id.
   */
  mergeActivists(newActivists: ActivistRecord[]): void {
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
  removeActivistsByIds(ids: number[]): void {
    const idsToRemove = new Set(ids)

    this.activists = this.activists.filter((activist) => {
      if (idsToRemove.has(activist.id)) {
        this.activistsByName.delete(activist.name)
        return false
      }
      return true
    })
  }

  /**
   * Get all activists as an array.
   */
  getActivists(): ActivistRecord[] {
    return this.activists
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
