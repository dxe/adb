import { describe, expect, it } from 'vitest'
import { decodeFirstEvent, encodeFirstEvent } from './filter-url-codecs'

describe('first event URL codec', () => {
  it('still reads date-only values from older links', () => {
    expect(decodeFirstEvent('-30..|null')).toEqual({
      date: { gte: { mode: 'relative', daysOffset: -30 }, orNull: true },
      types: undefined,
    })
    expect(
      encodeFirstEvent({
        date: { gte: { mode: 'relative', daysOffset: -30 } },
      }),
    ).toBe('-30..')
  })

  it('round-trips a date range with included types', () => {
    const value = {
      date: { gte: { mode: 'relative' as const, daysOffset: -30 } },
      types: {
        mode: 'include' as const,
        values: ['Outreach' as const, 'Campaign Action' as const],
      },
    }
    const encoded = encodeFirstEvent(value)
    expect(encoded).toBe('-30..;outreach,campaign-action')
    expect(decodeFirstEvent(encoded)).toEqual({
      date: { gte: { mode: 'relative', daysOffset: -30 }, lt: undefined },
      types: value.types,
    })
  })

  it('round-trips excluded types with no date range', () => {
    const value = {
      types: { mode: 'exclude' as const, values: ['Meeting' as const] },
    }
    expect(encodeFirstEvent(value)).toBe(';~meeting')
    expect(decodeFirstEvent(';~meeting')).toEqual({
      date: undefined,
      types: value.types,
    })
  })

  it('keeps "or none" alongside types', () => {
    expect(decodeFirstEvent('null;outreach')).toEqual({
      date: { orNull: true },
      types: { mode: 'include', values: ['Outreach'] },
    })
  })

  it('ignores unknown types', () => {
    expect(decodeFirstEvent(';outreach,bogus')?.types).toEqual({
      mode: 'include',
      values: ['Outreach'],
    })
    expect(decodeFirstEvent(';bogus')).toBeUndefined()
  })
})
