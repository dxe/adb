import { describe, expect, it } from 'vitest'
import {
  matchesAssignedToFilter,
  resolveDateBound,
  toApiDateRangeFilter,
  toApiFirstEvent,
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
      toApiDateRangeFilter(
        {
          gte: { mode: 'relative', daysOffset: -360 },
          lt: { mode: 'relative', daysOffset: 0 },
        },
        referenceDate,
      ),
    ).toEqual({
      date: { gte: '2025-03-21', lt: '2026-03-16' },
      or_null: undefined,
    })
  })

  it('puts or_null beside the date range', () => {
    const referenceDate = new Date('2026-03-16T23:30:00-07:00')

    expect(
      toApiDateRangeFilter(
        { lt: { mode: 'relative', daysOffset: -1 }, orNull: true },
        referenceDate,
      ),
    ).toEqual({ date: { gte: undefined, lt: '2026-03-15' }, or_null: true })
    expect(toApiDateRangeFilter({ orNull: true }, referenceDate)).toEqual({
      date: undefined,
      or_null: true,
    })
    expect(toApiDateRangeFilter({}, referenceDate)).toBeUndefined()
  })

  it('nests the first event date range and types', () => {
    const context = {
      chapterId: 1,
      userId: 1,
      referenceDate: new Date('2026-03-16T23:30:00-07:00'),
    }

    expect(
      toApiFirstEvent(
        {
          date: { gte: { mode: 'relative', daysOffset: -1 } },
          types: { mode: 'exclude', values: ['Meeting'] },
          orNull: true,
        },
        context,
      ),
    ).toEqual({
      first_event: {
        date: { gte: '2026-03-15', lt: undefined },
        types: { mode: 'exclude', values: ['Meeting'] },
        or_null: true,
      },
    })
    expect(
      toApiFirstEvent({ types: { mode: 'include', values: [] } }, context),
    ).toEqual({ first_event: undefined })
  })
})
