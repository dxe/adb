import { describe, expect, it } from 'vitest'
import { guardTrailingDot, withoutTrailingDotGuard } from './trailing-dot-guard'

/** Mirrors how nuqs renders a query string, so we assert on real URL text. */
const render = (search: URLSearchParams) =>
  '?' +
  Array.from(search.entries())
    .map(([key, value]) => `${key}=${value}`)
    .join('&')

describe('guardTrailingDot', () => {
  it('appends a guard when the last value ends in a dot', () => {
    const search = new URLSearchParams([
      ['columns', 'email'],
      ['lastEvent', '20260101..'],
    ])
    expect(render(guardTrailingDot(search))).toBe(
      '?columns=email&lastEvent=20260101..&=',
    )
  })

  it('leaves the params untouched when nothing ends in a dot', () => {
    const search = new URLSearchParams([['lastEvent', '20260101..20260601']])
    expect(guardTrailingDot(search)).toBe(search)
  })

  it('guards a dot-ending int range', () => {
    const search = new URLSearchParams([['totalEvents', '5..']])
    expect(render(guardTrailingDot(search))).toBe('?totalEvents=5..&=')
  })

  it('drops a stale guard once the dotted filter is gone', () => {
    const search = new URLSearchParams([
      ['lastEvent', '20260101..20260601'],
      ['', ''],
    ])
    expect(render(guardTrailingDot(search))).toBe(
      '?lastEvent=20260101..20260601',
    )
  })

  it('does not double up an existing guard', () => {
    const search = new URLSearchParams([
      ['lastEvent', '20260101..'],
      ['', ''],
    ])
    expect(render(guardTrailingDot(search))).toBe('?lastEvent=20260101..&=')
  })

  it('handles empty params', () => {
    const search = new URLSearchParams()
    expect(guardTrailingDot(search)).toBe(search)
  })

  it('round-trips to the same filter values a reader would see', () => {
    const guarded = guardTrailingDot(
      new URLSearchParams([['lastEvent', '20260101..']]),
    )
    const reparsed = new URLSearchParams(render(guarded))
    expect(reparsed.get('lastEvent')).toBe('20260101..')
  })

  it('strips the guard so guarded and unguarded urls compare equal', () => {
    // Which param renders last decides whether a URL is guarded, so the same
    // filters can produce a guarded URL one way and an unguarded one another.
    const guarded = new URLSearchParams([
      ['totalEvents', '1..'],
      ['', ''],
    ])
    const unguarded = new URLSearchParams([['totalEvents', '1..']])
    expect(render(withoutTrailingDotGuard(guarded))).toBe(
      render(withoutTrailingDotGuard(unguarded)),
    )
  })

  it('leaves params without a guard untouched', () => {
    const search = new URLSearchParams([['totalEvents', '1..']])
    expect(withoutTrailingDotGuard(search)).toBe(search)
  })
})
