import { describe, expect, it } from 'vitest'
import {
  matchesAssignedToFilter,
  resolveDateBound,
  toApiDateRange,
} from './filter-api-transform'

describe('matchesAssignedToFilter', () => {
  const USER_ID = 7
  const UNASSIGNED = 0

  it('matches everything when the filter is off', () => {
    expect(matchesAssignedToFilter(undefined, UNASSIGNED, USER_ID)).toBe(true)
  })

  it('holds when assigning to the user the filter names', () => {
    expect(matchesAssignedToFilter('me', USER_ID, USER_ID)).toBe(true)
    expect(matchesAssignedToFilter('42', 42, USER_ID)).toBe(true)
  })

  it('breaks when assigning to anyone else', () => {
    expect(matchesAssignedToFilter('me', 42, USER_ID)).toBe(false)
    expect(matchesAssignedToFilter('42', USER_ID, USER_ID)).toBe(false)
    expect(matchesAssignedToFilter('42', UNASSIGNED, USER_ID)).toBe(false)
  })

  it('holds for any assignee, but not for unassigning', () => {
    expect(matchesAssignedToFilter('any', 42, USER_ID)).toBe(true)
    expect(matchesAssignedToFilter('any', UNASSIGNED, USER_ID)).toBe(false)
  })
})

describe('filter-api-transform', () => {
  it('passes through absolute dates unchanged', () => {
    const referenceDate = new Date('2026-03-16T23:30:00-07:00')

    expect(
      resolveDateBound({ mode: 'absolute', date: '2025-01-15' }, referenceDate),
    ).toBe('2025-01-15')
  })

  it('resolves relative dates from the California day boundary', () => {
    const referenceDate = new Date('2026-03-16T23:30:00-07:00')

    expect(
      resolveDateBound({ mode: 'relative', daysOffset: 0 }, referenceDate),
    ).toBe('2026-03-16')
    expect(
      resolveDateBound({ mode: 'relative', daysOffset: -1 }, referenceDate),
    ).toBe('2026-03-15')
  })

  it('uses the same California-based reference date for both bounds', () => {
    const referenceDate = new Date('2026-03-16T23:30:00-07:00')

    expect(
      toApiDateRange(
        {
          gte: { mode: 'relative', daysOffset: -360 },
          lt: { mode: 'relative', daysOffset: 0 },
          orNull: true,
        },
        referenceDate,
      ),
    ).toEqual({
      gte: '2025-03-21',
      lt: '2026-03-16',
      or_null: true,
    })
  })
})
