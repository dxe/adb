/**
 * Keeps nuqs-written URLs from ending in a `.`.
 *
 * Our range filter syntax is `[gte]..[lt]`, so any open-ended range serializes
 * with trailing dots — `?lastEvent=20260101..` means "January 1 2026 or later".
 * nuqs renders the query string as `?k=v&k=v` in URLSearchParams insertion
 * order, so the URL ends in `..` whenever such a filter happens to land last.
 *
 * Link detectors read those dots as sentence punctuation and leave them out of
 * the link, so pasting the URL into Signal (and most other chat apps) produces
 * a link that silently drops the filter. Appending an empty-keyed param renders
 * as a trailing `&=`, which keeps the dots inside the link: linkify-it — the
 * detector Signal uses — only continues a URL past `..` when the next character
 * is alphanumeric or one of `%`, `/`, `&`.
 *
 * The guard param is inert: it parses back to an empty key, which every reader
 * here ignores. It is stripped and re-added on each update so it never lingers
 * once the open-ended filter is cleared.
 *
 * This only covers URLs nuqs writes, which is every URL built by interacting
 * with the filter UI. Loading an already-dotted link leaves the address bar
 * untouched until the next filter change.
 *
 * Keep in sync with `src/components/nav-hrefs.test.ts`, which enforces the same
 * rule on the hand-maintained hrefs in `shared/nav.json` — those bypass this
 * function entirely, so a change to what counts as a safe ending here needs the
 * matching change there.
 */
export const TRAILING_DOT_GUARD_KEY = ''

export function guardTrailingDot(search: URLSearchParams): URLSearchParams {
  const stripped = withoutTrailingDotGuard(search)
  const entries = Array.from(stripped.entries())

  if (!entries.at(-1)?.[1].endsWith('.')) {
    return stripped
  }

  const guarded = new URLSearchParams(entries)
  guarded.append(TRAILING_DOT_GUARD_KEY, '')
  return guarded
}

/**
 * Drops the guard param so two URLs can be compared on their meaningful params
 * regardless of whether either happens to carry one. Whether a URL ends up
 * guarded depends on which param renders last, so a hand-written link and the
 * one the app writes for the same filters may disagree.
 */
export function withoutTrailingDotGuard(
  search: URLSearchParams,
): URLSearchParams {
  if (!search.has(TRAILING_DOT_GUARD_KEY)) {
    return search
  }
  const stripped = new URLSearchParams(search)
  stripped.delete(TRAILING_DOT_GUARD_KEY)
  return stripped
}
