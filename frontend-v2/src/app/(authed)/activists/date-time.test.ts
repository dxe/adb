import { describe, expect, it } from 'vitest'
import { formatDateValueForActivists } from './date-time'

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
