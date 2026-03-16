import { describe, expect, it } from 'vitest'
import { resolveDateBound, toApiDateRange } from './filter-api-transform'

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
