import type { ActivistColumnName } from '@/lib/api'
import type { FilterState } from './query-state'
import { COLUMN_ORDER_BY_NAME } from './column-definitions'

// Maps filter keys to the column that should be auto-shown when the filter
// gets a value. Excludes searchAcrossChapters (handled separately) and
// training, includeHidden, nameSearch and prospect which have no single
// corresponding column.
const FILTER_COLUMN_MAP: Partial<
  Record<keyof FilterState, ActivistColumnName>
> = {
  lastEvent: 'last_event',
  firstEvent: 'first_event',
  totalEvents: 'total_events',
  interestDate: 'interest_date',
  totalInteractions: 'total_interactions',
  activistLevel: 'activist_level',
  source: 'source',
  assignedTo: 'assigned_to_name',
  followups: 'followup_date',
}

// Normalizes column selection. Ensures columns:
//  * include required columns
//  * are not duplicated
//  * are sorted according to their order in COLUMN_DEFINITIONS
export const normalizeColumns = (
  columns: ActivistColumnName[],
): ActivistColumnName[] => {
  const uniqueColumns = Array.from(new Set(columns))

  if (!uniqueColumns.includes('name')) {
    uniqueColumns.push('name')
  }

  return uniqueColumns.sort((a, b) => {
    const orderA = COLUMN_ORDER_BY_NAME.get(a) ?? Number.MAX_SAFE_INTEGER
    const orderB = COLUMN_ORDER_BY_NAME.get(b) ?? Number.MAX_SAFE_INTEGER
    return orderA - orderB
  })
}

/** Returns columns that should be added based on newly-set filter values. */
export const columnsForNewFilters = (
  prev: FilterState,
  next: FilterState,
): ActivistColumnName[] => {
  const newColumns: ActivistColumnName[] = []
  for (const [filterKey, columnName] of Object.entries(FILTER_COLUMN_MAP)) {
    const key = filterKey as keyof FilterState
    if (prev[key] === undefined && next[key] !== undefined) {
      newColumns.push(columnName)
    }
  }
  return newColumns
}

// Normalizes columns and ensures chapter_name is present if and only if
// searching across chapters. This centralizes the filter-dependent column logic.
export const normalizeColumnsForFilters = (
  columns: ActivistColumnName[],
  searchAcrossChapters: boolean,
): ActivistColumnName[] => {
  let adjustedColumns = [...columns]

  if (searchAcrossChapters) {
    if (!adjustedColumns.includes('chapter_name')) {
      adjustedColumns.unshift('chapter_name')
    }
  } else if (adjustedColumns.includes('chapter_name')) {
    adjustedColumns = adjustedColumns.filter((col) => col !== 'chapter_name')
  }

  return normalizeColumns(adjustedColumns)
}
