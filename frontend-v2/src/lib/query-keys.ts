import {
  API_PATH,
  type QueryActivistCountOptions,
  type QueryActivistOptions,
} from '@/lib/api'

// Query keys for TanStack Query.

export const activistKeys = {
  /** Every activist list query, whatever its filters. */
  lists: () => ['activists', 'list'] as const,
  list: (options: QueryActivistOptions) =>
    ['activists', 'list', options] as const,
  /** The name/id-only activist list used for autocomplete. */
  listBasic: () => [API_PATH.ACTIVIST_LIST_BASIC] as const,
  /** Every activist count query, whatever its filters. */
  counts: () => [API_PATH.ACTIVISTS_COUNT] as const,
  count: (options: QueryActivistCountOptions) =>
    [API_PATH.ACTIVISTS_COUNT, options] as const,
  /** Every cached activist detail, whichever activist. */
  details: () => ['activists', 'detail'] as const,
  detail: (activistId: number) => ['activists', 'detail', activistId] as const,
}
