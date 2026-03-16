import type { LocalDateYmd } from './date-time'

export type AbsoluteDateBound = {
  mode: 'absolute'
  date: LocalDateYmd
}

export type RelativeDateBound = {
  mode: 'relative'
  /** Signed day offset from today. */
  daysOffset: number
}

export type DateRangeBoundValue = AbsoluteDateBound | RelativeDateBound

export type DateRangeFilterValue = {
  gte?: DateRangeBoundValue
  lt?: DateRangeBoundValue
  orNull?: boolean
}

export type IntRangeFilterValue = {
  gte?: number
  lt?: number
}

export type IncludeExcludeFilterValue = {
  include: string[]
  exclude: string[]
}

export const ACTIVIST_LEVELS = [
  'Supporter',
  'Chapter Member',
  'Organizer',
  'Non-Local',
  'Global Network Member',
] as const

export type ActivistLevelValue = (typeof ACTIVIST_LEVELS)[number]

export type ActivistLevelFilterValue = {
  mode: 'include' | 'exclude'
  values: ActivistLevelValue[]
}

export type AssignedToFilterValue = 'me' | 'any' | `${number}`

export type FollowupsFilterValue = 'all' | 'due' | 'upcoming'

export type ProspectFilterValue = 'chapterMember' | 'organizer'
