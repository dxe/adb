type ActivistRecord = {
  name: string
  email?: string
  phone?: string
}

/**
 * Encapsulates basic activist data and provides lookup methods.
 */
export class ActivistRegistry {
  private byName: Map<string, ActivistRecord>
  private names: Set<string>
  private activists: ActivistRecord[]

  constructor(activists: ActivistRecord[]) {
    this.activists = activists
    this.byName = new Map(activists.map((a) => [a.name, a]))
    this.names = new Set(activists.map((a) => a.name))
  }

  hasEmail(name: string): boolean {
    return !!this.byName.get(name)?.email
  }

  hasPhone(name: string): boolean {
    return !!this.byName.get(name)?.phone
  }

  exists(name: string): boolean {
    return this.names.has(name)
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

// Like the Vue nameFilter: internal whitespace matches anything.
function nameFilter(text: string, input: string): boolean {
  const pattern = input
    .trim()
    .replace(/[.*+?^${}()|[\]\\]/g, '\\$&') // escape special regex chars
    .replace(/ +/g, '.*') // whitespace matches anything
  return new RegExp(pattern, 'i').test(text)
}
