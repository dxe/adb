import { z } from 'zod'

// Activist data structure returned by the API.
export const ActivistJSON = z.object({
  id: z.number().optional(),
  name: z.string().optional(),
  preferred_name: z.string().optional(),
  email: z.string().optional(),
  phone: z.string().optional(),
  pronouns: z.string().optional(),
  language: z.string().optional(),
  accessibility: z.string().optional(),
  dob: z.string().optional(),
  facebook: z.string().optional(),
  location: z.string().optional(),
  street_address: z.string().optional(),
  city: z.string().optional(),
  state: z.string().optional(),
  lat: z.number().optional(),
  lng: z.number().optional(),
  chapter_id: z.number().optional(),
  chapter_name: z.string().optional(),
  activist_level: z.string().optional(),
  source: z.string().optional(),
  hiatus: z.boolean().optional(),
  connector: z.string().optional(),
  training0: z.string().optional(),
  training1: z.string().optional(),
  training4: z.string().optional(),
  training5: z.string().optional(),
  training6: z.string().optional(),
  consent_quiz: z.string().optional(),
  training_protest: z.string().optional(),
  dev_application_date: z.string().optional(),
  dev_application_type: z.string().optional(),
  dev_quiz: z.string().optional(),
  dev_interest: z.string().optional(),
  cm_first_email: z.string().optional(),
  cm_approval_email: z.string().optional(),
  prospect_organizer: z.boolean().optional(),
  prospect_chapter_member: z.boolean().optional(),
  referral_friends: z.string().optional(),
  referral_apply: z.string().optional(),
  referral_outlet: z.string().optional(),
  interest_date: z.string().optional(),
  mpi: z.boolean().optional(),
  notes: z.string().optional(),
  vision_wall: z.string().optional(),
  mpp_requirements: z.string().optional(),
  voting_agreement: z.boolean().optional(),
  assigned_to: z.number().optional(),
  followup_date: z.string().optional(),
  first_event: z.string().optional(),
  first_event_name: z.string().optional(),
  last_event: z.string().optional(),
  last_event_name: z.string().optional(),
  total_events: z.number().optional(),
  last_action: z.string().optional(),
  months_since_last_action: z.number().optional(),
  total_points: z.number().optional(),
  active: z.boolean().optional(),
  status: z.string().optional(),
  last_connection: z.string().optional(),
  geo_circles: z.string().optional(),
  assigned_to_name: z.string().optional(),
  total_interactions: z.number().optional(),
  last_interaction_date: z.string().optional(),
})
export type ActivistJSON = z.infer<typeof ActivistJSON>

export const ActivistColumnName = z.enum(
  Object.keys(ActivistJSON.shape) as [string, ...string[]],
)
export type ActivistColumnName = z.infer<typeof ActivistColumnName>

const ActivistNameFilter = z.object({
  name_contains: z.string().optional(),
})

const DateRangeFilter = z.object({
  gte: z.string().optional(),
  lt: z.string().optional(),
  or_null: z.boolean().optional(),
})

const IntRangeFilter = z.object({
  gte: z.number().optional(),
  lt: z.number().optional(),
})

const ActivistLevelFilter = z.object({
  mode: z.enum(['include', 'exclude']).optional(),
  values: z.array(z.string()).optional(),
})

const SourceFilter = z.object({
  contains_any: z.array(z.string()).optional(),
  not_contains_any: z.array(z.string()).optional(),
})

const TrainingFilter = z.object({
  completed: z.array(z.string()).optional(),
  not_completed: z.array(z.string()).optional(),
})

const QueryActivistFilters = z.object({
  chapter_id: z.number().optional(),
  name: ActivistNameFilter.optional(),
  last_event: DateRangeFilter.optional(),
  include_hidden: z.boolean().optional(),
  activist_level: ActivistLevelFilter.optional(),
  interest_date: DateRangeFilter.optional(),
  first_event: DateRangeFilter.optional(),
  total_events: IntRangeFilter.optional(),
  total_interactions: IntRangeFilter.optional(),
  source: SourceFilter.optional(),
  training: TrainingFilter.optional(),
  assigned_to: z.number().optional(),
  followups: z.enum(['all', 'due', 'upcoming']).optional(),
  prospect: z.enum(['chapter_member', 'organizer']).optional(),
})

const ActivistSortColumn = z.object({
  column_name: ActivistColumnName,
  desc: z.boolean(),
})

const ActivistSortOptions = z.object({
  sort_columns: z.array(ActivistSortColumn),
})

export const QueryActivistOptions = z.object({
  columns: z.array(ActivistColumnName),
  filters: QueryActivistFilters,
  sort: ActivistSortOptions.optional(),
  after: z.string().optional(),
})
export type QueryActivistOptions = z.infer<typeof QueryActivistOptions>

const QueryActivistPagination = z.object({
  next_cursor: z.string(),
})

export const QueryActivistResult = z.object({
  activists: z.array(ActivistJSON),
  pagination: QueryActivistPagination,
})
export type QueryActivistResult = z.infer<typeof QueryActivistResult>
