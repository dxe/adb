import type { LocalDateYmd } from './date-time'
import { EVENT_TYPES } from '../events/event-form-schema'

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

// Event types the first event filter offers. Unlike the event form, includes
// connections, since they count as attended events.
export const FIRST_EVENT_TYPES = [...EVENT_TYPES, 'Connection'] as const

export type FirstEventTypeValue = (typeof FIRST_EVENT_TYPES)[number]

export type EventTypeFilterValue = {
  mode: 'include' | 'exclude'
  values: FirstEventTypeValue[]
}

export type FirstEventFilterValue = {
  date?: DateRangeFilterValue
  types?: EventTypeFilterValue
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

// Ways an activist may ask to be contacted. Keep in sync with
// `contactMethods` in pkg/activists/contact_method.go.
export const CONTACT_METHODS = [
  'Signal',
  'Text / SMS',
  'Phone (Call)',
  'Email',
  'WhatsApp',
  'Telegram',
  'Instagram',
  'Facebook',
] as const

export type ActivistLevelFilterValue = {
  mode: 'include' | 'exclude'
  values: ActivistLevelValue[]
}

export type AssignedToFilterValue = 'me' | 'any' | `${number}`

export type FollowupsFilterValue = 'all' | 'due' | 'upcoming'

export type ProspectFilterValue = 'chapterMember' | 'organizer'
