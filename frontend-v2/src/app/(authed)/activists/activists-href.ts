import { createSerializer } from 'nuqs/server'
import { TRAILING_DOT_GUARD_KEY } from '@/lib/trailing-dot-guard'
import {
  ACTIVIST_QUERY_STATE_PARSERS,
  ACTIVIST_QUERY_URL_KEYS,
  type ParsedActivistQueryParams,
} from './search-params'

const serializeActivistQuery = createSerializer(ACTIVIST_QUERY_STATE_PARSERS, {
  urlKeys: ACTIVIST_QUERY_URL_KEYS,
})

/** Builds an `/activists` path (no basePath) from typed filters, columns and sort. */
export function buildActivistsHref(
  query: Partial<ParsedActivistQueryParams>,
): string {
  // Appends the guard directly: a URLSearchParams round-trip would percent-encode
  // the `,`, `~` and `|` that nuqs leaves readable.
  const href = serializeActivistQuery('/activists', query)
  return href.endsWith('.') ? `${href}&${TRAILING_DOT_GUARD_KEY}=` : href
}
