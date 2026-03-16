import type { SortColumn } from './query-state'

/** Builds URL param value for sort state. Returns undefined only when sort is empty. */
export const buildSortParam = (sort: SortColumn[]): string | undefined => {
  if (sort.length === 0) return undefined

  return sort.map((s) => (s.desc ? `-${s.column}` : s.column)).join(',')
}
