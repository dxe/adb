import { describe, expect, it } from 'vitest'
import { formatDateRange } from './date-range-filter'
import type { DateRangeFilterValue } from '../filter-types'

const relative = (daysOffset: number) => ({
  mode: 'relative' as const,
  daysOffset,
})

describe('formatDateRange', () => {
  describe('relative ranges', () => {
    it.each([
      {
        label: 'gte negative, lt negative',
        value: { gte: relative(-180), lt: relative(-30) },
        expected: '6 months ago – 1 month ago',
      },
      {
        label: 'gte negative, lt zero',
        value: { gte: relative(-7), lt: relative(0) },
        expected: '1 week ago – today',
      },
      {
        label: 'gte negative, lt positive',
        value: { gte: relative(-7), lt: relative(14) },
        expected: '1 week ago – 2 weeks from now',
      },
      {
        label: 'gte negative, lt null',
        value: { gte: relative(-45) },
        expected: 'Last 45 days',
      },
      {
        label: 'gte zero, lt negative',
        value: { gte: relative(0), lt: relative(-7) },
        expected: 'No matching dates',
      },
      {
        label: 'gte zero, lt zero',
        value: { gte: relative(0), lt: relative(0) },
        expected: 'No matching dates',
      },
      {
        label: 'gte zero, lt positive',
        value: { gte: relative(0), lt: relative(14) },
        expected: 'today – 2 weeks from now',
      },
      {
        label: 'gte zero, lt null',
        value: { gte: relative(0) },
        expected: 'Today onward',
      },
      {
        label: 'gte positive, lt negative',
        value: { gte: relative(14), lt: relative(-7) },
        expected: 'No matching dates',
      },
      {
        label: 'gte positive, lt zero',
        value: { gte: relative(14), lt: relative(0) },
        expected: 'No matching dates',
      },
      {
        label: 'gte positive, lt positive',
        value: { gte: relative(7), lt: relative(14) },
        expected: '1 week from now – 2 weeks from now',
      },
      {
        label: 'gte positive, lt null',
        value: { gte: relative(14) },
        expected: 'On or after 2 weeks from now',
      },
      {
        label: 'gte null, lt negative',
        value: { lt: relative(-180) },
        expected: 'Over 6 months ago',
      },
      {
        label: 'gte null, lt zero',
        value: { lt: relative(0) },
        expected: 'Before today',
      },
      {
        label: 'gte null, lt positive',
        value: { lt: relative(7) },
        expected: 'Before 1 week from now',
      },
      {
        label: 'gte null, lt null',
        value: {},
        expected: undefined,
      },
    ])(
      '$label',
      ({
        value,
        expected,
      }: {
        value: DateRangeFilterValue
        expected?: string
      }) => {
        expect(formatDateRange(value)).toBe(expected)
      },
    )

    it('uses singular units where needed', () => {
      const value: DateRangeFilterValue = {
        gte: relative(-30),
      }
      expect(formatDateRange(value)).toBe('Last 1 month')
    })

    it('returns no matches for reversed negative closed ranges', () => {
      const value: DateRangeFilterValue = {
        gte: relative(-30),
        lt: relative(-180),
      }
      expect(formatDateRange(value)).toBe('No matching dates')
    })

    it('returns no matches for reversed positive closed ranges', () => {
      const value: DateRangeFilterValue = {
        gte: relative(14),
        lt: relative(7),
      }
      expect(formatDateRange(value)).toBe('No matching dates')
    })
  })

  describe('absolute ranges', () => {
    it('closed range in the same year as today omits years', () => {
      const currentYear = new Date().getFullYear()
      const value: DateRangeFilterValue = {
        gte: { mode: 'absolute', date: `${currentYear}-01-01` },
        lt: { mode: 'absolute', date: `${currentYear}-06-15` },
      }
      expect(formatDateRange(value)).toBe('Jan 1 – Jun 15')
    })

    it('closed range spanning different years includes years', () => {
      const value: DateRangeFilterValue = {
        gte: { mode: 'absolute', date: '2024-03-01' },
        lt: { mode: 'absolute', date: '2025-09-15' },
      }
      expect(formatDateRange(value)).toBe('Mar 1, 2024 – Sep 15, 2025')
    })

    it('lower bound only: "On or after"', () => {
      const value: DateRangeFilterValue = {
        gte: { mode: 'absolute', date: '2024-06-01' },
      }
      expect(formatDateRange(value)).toBe('On or after Jun 1, 2024')
    })

    it('upper bound only: "Before"', () => {
      const value: DateRangeFilterValue = {
        lt: { mode: 'absolute', date: '2025-01-01' },
      }
      expect(formatDateRange(value)).toBe('Before Jan 1, 2025')
    })

    it('closed range with start on or after end reports no matches', () => {
      const value: DateRangeFilterValue = {
        gte: { mode: 'absolute', date: '2025-06-15' },
        lt: { mode: 'absolute', date: '2025-06-15' },
      }
      expect(formatDateRange(value)).toBe('No matching dates')
    })
  })

  describe('with orNull', () => {
    it('range with orNull appends "or none"', () => {
      const value: DateRangeFilterValue = {
        gte: { mode: 'relative', daysOffset: -180 },
        orNull: true,
      }
      expect(formatDateRange(value)).toBe('Last 6 months or none')
    })

    it('orNull only shows "None"', () => {
      const value: DateRangeFilterValue = {
        orNull: true,
      }
      expect(formatDateRange(value)).toBe('None')
    })

    it('absolute range with orNull', () => {
      const value: DateRangeFilterValue = {
        lt: { mode: 'absolute', date: '2025-01-01' },
        orNull: true,
      }
      expect(formatDateRange(value)).toBe('Before Jan 1, 2025 or none')
    })
  })

  describe('empty/undefined', () => {
    it('returns undefined for undefined input', () => {
      expect(formatDateRange(undefined)).toBeUndefined()
    })

    it('returns undefined for empty object', () => {
      expect(formatDateRange({})).toBeUndefined()
    })

    it('returns undefined for orNull: false', () => {
      expect(formatDateRange({ orNull: false })).toBeUndefined()
    })
  })
})
