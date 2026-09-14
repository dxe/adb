import { describe, expect, it } from 'vitest'
import { buildQueryOptions } from './filter-api-query'
import type { FilterState } from './query-state'

const FILTERS: FilterState = {
  searchAcrossChapters: false,
  nameSearch: '',
  includeHidden: true,
}

const baseInput = {
  filters: FILTERS,
  chapterId: 1,
  userId: 2,
  referenceDate: new Date('2026-01-01T00:00:00Z'),
  sort: [],
}

describe('buildQueryOptions', () => {
  it("requests the 'hidden' column even though it is never user-selected", () => {
    const options = buildQueryOptions({
      ...baseInput,
      selectedColumns: ['name'],
    })

    expect(options.shape.columns).toContain('hidden')
  })

  it("does not duplicate 'hidden' if it is already requested", () => {
    const options = buildQueryOptions({
      ...baseInput,
      selectedColumns: ['name', 'hidden'],
    })

    expect(
      options.shape.columns.filter((col) => col === 'hidden'),
    ).toHaveLength(1)
  })
})
