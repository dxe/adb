import { describe, expect, it } from 'vitest'
import {
  buildQueryOptions,
  buildSelectionQueryOptions,
  withVisibleColumns,
} from './filter-api-query'
import type { QueryActivistOptions } from '@/lib/api'
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

describe('buildSelectionQueryOptions', () => {
  const options: QueryActivistOptions = {
    shape: {
      columns: ['name', 'email'],
      filters: {
        chapter_id: 1,
        include_hidden: true,
        name: { name_contains: 'ali' },
        assigned_to: 7,
      },
      sort: { sort_columns: [{ column_name: 'name', desc: false }] },
    },
  }

  it('restricts the query to the given activists', () => {
    const selection = buildSelectionQueryOptions(options, [10, 20])

    expect(selection.shape.filters.ids).toEqual([10, 20])
  })

  it('keeps the columns and sort so the export matches the table', () => {
    const selection = buildSelectionQueryOptions(options, [10])

    expect(selection.shape.columns).toEqual(options.shape.columns)
    expect(selection.shape.sort).toEqual(options.shape.sort)
  })

  // The user picked these rows, so rows that have since stopped matching a
  // filter still belong in the result. Only the chapter (which the server
  // authorizes against) and include_hidden are kept.
  it('drops the other filters', () => {
    const selection = buildSelectionQueryOptions(options, [10])

    expect(selection.shape.filters).toEqual({
      chapter_id: 1,
      include_hidden: true,
      ids: [10],
    })
  })
})

describe('withVisibleColumns', () => {
  // buildQueryOptions asks for these on the table's behalf, not the user's.
  it("drops the table-only 'id' and 'hidden' columns", () => {
    const options = buildQueryOptions({
      ...baseInput,
      selectedColumns: ['name', 'email'],
    })

    expect(options.shape.columns).toEqual(
      expect.arrayContaining(['id', 'hidden']),
    )
    expect(
      withVisibleColumns(options, ['name', 'email']).shape.columns,
    ).toEqual(['name', 'email'])
  })

  it("keeps 'id' when the user chose to see it", () => {
    const options = buildQueryOptions({
      ...baseInput,
      selectedColumns: ['name', 'id'],
    })

    expect(withVisibleColumns(options, ['name', 'id']).shape.columns).toEqual([
      'name',
      'id',
    ])
  })
})
