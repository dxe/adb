type ActivistRecord = {
  name: string
  email?: string
  phone?: string
}

/**
 * Encapsulates basic activist data and provides lookup methods.
 */
export class ActivistRegistry {
  private activists: ActivistRecord[]
  private activistsByName: Map<string, ActivistRecord>

  constructor(activists: ActivistRecord[]) {
    this.activists = activists
    this.activistsByName = new Map(activists.map((a) => [a.name, a]))
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
