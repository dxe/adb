import { ActivistColumnName } from '@/lib/api'
import type {
  ActivistLevelFilterValue,
  AssignedToFilterValue,
  DateRangeFilterValue,
  FollowupsFilterValue,
  IncludeExcludeFilterValue,
  IntRangeFilterValue,
  ProspectFilterValue,
} from './filter-types'

export type FilterState = {
  searchAcrossChapters: boolean
  nameSearch: string
  includeHidden: boolean
  lastEvent?: DateRangeFilterValue
  interestDate?: DateRangeFilterValue
  firstEvent?: DateRangeFilterValue
  totalEvents?: IntRangeFilterValue
  totalInteractions?: IntRangeFilterValue
  activistLevel?: ActivistLevelFilterValue
  source?: IncludeExcludeFilterValue
  training?: IncludeExcludeFilterValue
  assignedTo?: AssignedToFilterValue
  followups?: FollowupsFilterValue
  prospect?: ProspectFilterValue
}

export type SortColumn = {
  column: ActivistColumnName
  desc: boolean
}

export type ActivistsQueryState = {
  filters: FilterState
  selectedColumns: ActivistColumnName[]
  sort: SortColumn[]
}

export const DEFAULT_SORT: SortColumn[] = [{ column: 'name', desc: false }]
