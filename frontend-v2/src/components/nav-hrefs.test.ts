import { describe, expect, it } from 'vitest'
import navbarData from '$shared/nav.json'
import type { NavbarData } from '$shared/nav-access'

/**
 * These hrefs are hand-maintained and never pass through `guardTrailingDot`,
 * so the rule it applies to app-written URLs is asserted here instead: a URL
 * must not end in `.`. See `src/lib/trailing-dot-guard.ts` for why, and for the
 * `&=` fix. Keep the two in sync.
 */
const hrefs = (navbarData as NavbarData).items.flatMap((group) =>
  group.items.map((item) => item.href),
)

describe('nav.json hrefs', () => {
  it('finds hrefs to check', () => {
    // Guards against the traversal silently matching nothing if nav.json's
    // shape changes, which would make the assertions below vacuous.
    expect(hrefs.length).toBeGreaterThan(20)
    expect(hrefs.every((href) => href.startsWith('/'))).toBe(true)
  })

  it.each(hrefs)('%s does not end in a dot', (href) => {
    expect(href.endsWith('.')).toBe(false)
  })
})
