import { describe, expect, it } from 'vitest'
import { navbarData } from './nav-data'

/**
 * These hrefs don't pass through `guardTrailingDot` (activist ones are guarded
 * by `buildActivistsHref`, the rest are hand-written), so the rule it applies to
 * app-written URLs is asserted here instead: a URL must not end in `.`. See `src/lib/trailing-dot-guard.ts` for why, and for the
 * `&=` fix. Keep the two in sync.
 */
const hrefs = navbarData.items.flatMap((group) =>
  group.items.map((item) => item.href),
)

describe('nav-data hrefs', () => {
  it('finds hrefs to check', () => {
    // Guards against the traversal silently matching nothing if nav-data's
    // shape changes, which would make the assertions below vacuous.
    expect(hrefs.length).toBeGreaterThan(20)
    expect(hrefs.every((href) => href.startsWith('/'))).toBe(true)
  })

  it.each(hrefs)('%s does not end in a dot', (href) => {
    expect(href.endsWith('.')).toBe(false)
  })
})
