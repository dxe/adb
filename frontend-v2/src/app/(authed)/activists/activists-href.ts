import { createSerializer } from 'nuqs/server'
import { guardTrailingDot } from '@/lib/trailing-dot-guard'
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
  const [path, search = ''] = serializeActivistQuery('/activists', query).split(
    '?',
  )
  const guarded = guardTrailingDot(new URLSearchParams(search)).toString()
  return guarded ? `${path}?${guarded}` : path
}
