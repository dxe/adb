import { describe, expect, it } from 'vitest'
import { buildActivistsHref } from './activists-href'

describe('buildActivistsHref', () => {
  it('returns the bare path when there is no query', () => {
    expect(buildActivistsHref({})).toBe('/activists')
  })

  it('serializes filters, columns and sort with the URL keys', () => {
    const href = buildActivistsHref({
      activistLevel: { mode: 'include', values: ['Supporter'] },
      lastEvent: { lt: { mode: 'relative', daysOffset: -360 }, orNull: true },
      columns: ['email', 'notes'],
      sort: [{ column: 'last_event', desc: true }],
    })
    const params = new URL(href, 'http://x').searchParams
    expect(params.get('level')).toBe('supporter')
    expect(params.get('lastEvent')).toBe('..-360|null')
    expect(params.get('columns')).toBe('email,notes')
    expect(params.get('sort')).toBe('-last_event')
  })

  it('guards a trailing open-ended range so link detectors keep the dots', () => {
    const href = buildActivistsHref({ totalEvents: { gte: 1 } })
    expect(href.endsWith('..&=')).toBe(true)
  })
})
