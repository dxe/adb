import { describe, expect, it } from 'vitest'
import {
  formatDateValueForActivists,
  isDateWithinLastMonths,
} from './date-time'

describe('formatDateValueForActivists', () => {
  it('treats date-only strings as activists local dates', () => {
    expect(formatDateValueForActivists('2026-01-23')).toBe('Jan 23, 2026')
  })

  it('treats timezone-less datetime strings as activists local datetimes', () => {
    expect(formatDateValueForActivists('2026-01-23T00:00:00')).toBe(
      'Jan 23, 2026',
    )
    expect(formatDateValueForActivists('2026-01-23 00:00:00')).toBe(
      'Jan 23, 2026',
    )
  })

  it('preserves explicit timezone offsets', () => {
    expect(formatDateValueForActivists('2026-01-23T00:00:00Z')).toBe(
      'Jan 22, 2026',
    )
  })
})

describe('isDateWithinLastMonths', () => {
  const reference = new Date('2026-04-15T12:00:00Z')

  it('accepts dates inside the window', () => {
    expect(isDateWithinLastMonths('2026-04-15', 3, reference)).toBe(true)
    expect(isDateWithinLastMonths('2026-01-16', 3, reference)).toBe(true)
  })

  it('rejects dates outside the window', () => {
    expect(isDateWithinLastMonths('2026-01-14', 3, reference)).toBe(false)
    expect(isDateWithinLastMonths('2025-04-15', 3, reference)).toBe(false)
  })

  it('rejects unparseable dates', () => {
    expect(isDateWithinLastMonths('not a date', 3, reference)).toBe(false)
  })
})
