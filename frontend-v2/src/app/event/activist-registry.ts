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

  getActivist(name: string): {
    exists: boolean
    hasEmail: boolean
    hasPhone: boolean
  } {
    const activist = this.activistsByName.get(name)
    return {
      exists: !!activist,
      hasEmail: !!activist?.email,
      hasPhone: !!activist?.phone,
    }
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

function nameFilter(text: string, input: string): boolean {
  const pattern = input
    .trim()
    .replace(/[.*+?^${}()|[\]\\]/g, '\\$&') // escape special regex chars
    .replace(/ +/g, '.*') // whitespace matches anything
  return new RegExp(pattern, 'i').test(text)
}
