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
})
export type ActivistJSON = z.infer<typeof ActivistJSON>

export const ActivistColumnName = z.enum(
  Object.keys(ActivistJSON.shape) as [string, ...string[]],
)
export type ActivistColumnName = z.infer<typeof ActivistColumnName>

const ActivistNameFilter = z.object({
  name_contains: z.string().optional(),
})

const LastEventFilter = z.object({
  last_event_lt: z.string().optional(), // ISO date string (YYYY-MM-DD)
  last_event_gte: z.string().optional(), // ISO date string (YYYY-MM-DD)
})

const QueryActivistFilters = z.object({
  chapter_id: z.number().optional(),
  name: ActivistNameFilter.optional(),
  last_event: LastEventFilter.optional(),
  include_hidden: z.boolean().optional(),
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
