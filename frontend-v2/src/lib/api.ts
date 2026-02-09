import ky, { HTTPError, KyInstance } from 'ky'
import { z } from 'zod'

export const API_PATH = {
  STATIC_RESOURCE_HASH: 'static_resources_hash',
  ACTIVIST_NAMES_GET: 'activist_names/get',
  ACTIVIST_LIST_BASIC: 'activist/list_basic',
  ACTIVISTS_SEARCH: 'api/activists',
  USER_ME: 'user/me',
  CSRF_TOKEN: 'api/csrf-token',
  CHAPTER_LIST: 'chapter/list',
  USERS: 'api/users',
  EVENT_GET: 'event/get',
  EVENT_SAVE: 'event/save',
}

export const StaticResourcesHashResp = z.object({
  hash: z.string(),
})

export const Role = z.enum(['admin', 'organizer', 'attendance', 'non-sfbay'])
export type Role = z.infer<typeof Role>

const AuthedUserResp = z.object({
  user: z.object({
    ChapterID: z.number(),
    ChapterName: z.string(),
    Disabled: z.boolean(),
    Email: z.string(),
    ID: z.number(),
    Name: z.string(),
    Roles: z.array(Role),
  }),
  mainRole: Role,
})

const CsrfTokenResp = z.object({
  status: z.literal('success'),
  csrfToken: z.string(),
})

const ChapterListResp = z.object({
  chapters: z.array(
    z.object({
      ChapterID: z.number(),
      Name: z.string(),
    }),
  ),
})

const RolesSchema = z
  .array(Role)
  .nullable()
  .transform((roles) => roles ?? [])

const UserSchema = z.object({
  id: z.number(),
  email: z.string(),
  name: z.string(),
  disabled: z.boolean(),
  roles: RolesSchema,
  chapter_id: z.number(),
})
export type User = z.infer<typeof UserSchema>
export type UserWithoutId = Omit<User, 'id'>

const UserListResp = z.object({
  users: z.array(UserSchema),
})

const UserSaveResp = z.object({
  status: z.literal('success'),
  user: UserSchema,
})

const UserGetResp = z.object({
  user: UserSchema,
})

export const ActivistNamesResp = z.object({
  activist_names: z.array(z.string()),
})

export const ActivistListBasicResp = z.object({
  activists: z.array(
    z.object({
      id: z.number(),
      name: z.string(),
      email: z.boolean(),
      phone: z.boolean(),
      last_updated: z.number(), // Unix timestamp in seconds
    }),
  ),
  hidden_ids: z.array(z.number()),
})

export type ActivistListBasic = z.infer<typeof ActivistListBasicResp>

// Activists Search API Types
export const ActivistColumnName = z.enum([
  'id',
  'name',
  'preferred_name',
  'email',
  'phone',
  'pronouns',
  'language',
  'accessibility',
  'dob',
  'facebook',
  'location',
  'street_address',
  'city',
  'state',
  'lat',
  'lng',
  'chapter_id',
  'chapter_name',
  'activist_level',
  'source',
  'hiatus',
  'connector',
  'training0',
  'training1',
  'training4',
  'training5',
  'training6',
  'consent_quiz',
  'training_protest',
  'dev_application_date',
  'dev_application_type',
  'dev_quiz',
  'dev_interest',
  'cm_first_email',
  'cm_approval_email',
  'prospect_organizer',
  'prospect_chapter_member',
  'referral_friends',
  'referral_apply',
  'referral_outlet',
  'interest_date',
  'mpi',
  'notes',
  'vision_wall',
  'mpp_requirements',
  'voting_agreement',
  'assigned_to',
  'followup_date',
  'first_event',
  'first_event_name',
  'last_event',
  'last_event_name',
  'total_events',
])
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
  mpi: z.number().optional(),
  notes: z.string().optional(),
  vision_wall: z.string().optional(),
  mpp_requirements: z.string().optional(),
  voting_agreement: z.string().optional(),
  assigned_to: z.string().optional(),
  followup_date: z.string().optional(),
  first_event: z.string().optional(),
  first_event_name: z.string().optional(),
  last_event: z.string().optional(),
  last_event_name: z.string().optional(),
  total_events: z.number().optional(),
})
export type ActivistJSON = z.infer<typeof ActivistJSON>

const QueryActivistPagination = z.object({
  next_cursor: z.string(),
})

export const QueryActivistResult = z.object({
  activists: z.array(ActivistJSON),
  pagination: QueryActivistPagination,
})
export type QueryActivistResult = z.infer<typeof QueryActivistResult>

const EventGetResp = z.object({
  event: z.object({
    event_name: z.string(),
    event_type: z.string(),
    event_date: z.string(),
    attendees: z.array(z.string()).nullable(),
    suppress_survey: z.boolean(),
  }),
})

export type EventData = z.infer<typeof EventGetResp>['event']

interface SaveEventParams {
  event_id: number
  event_name: string
  event_date: string
  event_type: string
  added_attendees: string[]
  deleted_attendees: string[]
  suppress_survey: boolean
}

const EventSaveResp = z.object({
  status: z.literal('success'),
  redirect: z.string().optional(),
  attendees: z.array(z.string()).nullish(),
})

const ApiErrorResp = z.object({
  status: z.literal('error'),
  message: z.string(),
})

export class ApiClient {
  private client: KyInstance

  constructor(cookies?: string) {
    this.client = ky.extend({
      prefixUrl:
        typeof window === 'undefined'
          ? process.env.NEXT_PUBLIC_API_BASE_URL
          : '/',
      headers: cookies
        ? {
            Cookie: cookies,
          }
        : undefined,
    })
  }

  private async handleKyError(err: unknown): Promise<never> {
    if (err instanceof HTTPError) {
      const parsed = ApiErrorResp.safeParse(
        await err.response.json().catch(() => null),
      )
      if (parsed.success) {
        throw new Error(parsed.data.message)
      }
    }
    throw err
  }

  getAuthedUser = async () => {
    try {
      const resp = await this.client.get(API_PATH.USER_ME).json()
      return AuthedUserResp.parse(resp)
    } catch (err) {
      console.error(`Error fetching authed user: ${err}`)
      return {
        user: null,
        mainRole: null,
      }
    }
  }

  getStaticResourceHash = async () => {
    const resp = await this.client.get(API_PATH.STATIC_RESOURCE_HASH).json()
    return StaticResourcesHashResp.parse(resp)
  }

  async fetchCsrfToken() {
    try {
      const resp = await this.client.get(API_PATH.CSRF_TOKEN).json()
      return CsrfTokenResp.parse(resp).csrfToken
    } catch (err) {
      return this.handleKyError(err)
    }
  }

  private getCsrfToken(): string | undefined {
    if (typeof document === 'undefined') return undefined

    const metaToken = document
      .querySelector('meta[name="csrf-token"]')
      ?.getAttribute('content')
    return metaToken ? metaToken : undefined
  }

  getActivistNames = async () => {
    const resp = await this.client.get(API_PATH.ACTIVIST_NAMES_GET).json()
    return ActivistNamesResp.parse(resp)
  }

  getActivistListBasic = async (modifiedSince?: string) => {
    const options = modifiedSince
      ? { searchParams: { modified_since: modifiedSince } }
      : {}

    const resp = await this.client
      .get(API_PATH.ACTIVIST_LIST_BASIC, options)
      .json()
    return ActivistListBasicResp.parse(resp)
  }

  searchActivists = async (options: QueryActivistOptions) => {
    try {
      const resp = await this.client
        .post(API_PATH.ACTIVISTS_SEARCH, { json: options })
        .json()
      return QueryActivistResult.parse(resp)
    } catch (err) {
      return this.handleKyError(err)
    }
  }

  getChapterList = async () => {
    try {
      const resp = await this.client.get(API_PATH.CHAPTER_LIST).json()
      return ChapterListResp.parse(resp).chapters
    } catch (err) {
      return this.handleKyError(err)
    }
  }

  getUsers = async () => {
    try {
      const resp = await this.client.get(API_PATH.USERS).json()
      return UserListResp.parse(resp).users
    } catch (err) {
      return this.handleKyError(err)
    }
  }

  getUser = async (userId: number) => {
    try {
      const resp = await this.client.get(`${API_PATH.USERS}/${userId}`).json()
      return UserGetResp.parse(resp).user
    } catch (err) {
      return this.handleKyError(err)
    }
  }

  createUser = async (payload: UserWithoutId) => {
    try {
      const csrfToken = this.getCsrfToken()
      const resp = await this.client
        .post(API_PATH.USERS, {
          json: payload,
          headers: { 'X-CSRF-Token': csrfToken },
        })
        .json()
      return UserSaveResp.parse(resp).user
    } catch (err) {
      return this.handleKyError(err)
    }
  }

  updateUser = async (payload: User) => {
    try {
      const csrfToken = this.getCsrfToken()
      const resp = await this.client
        .put(`${API_PATH.USERS}/${payload.id}`, {
          json: payload,
          headers: { 'X-CSRF-Token': csrfToken },
        })
        .json()
      return UserSaveResp.parse(resp).user
    } catch (err) {
      return this.handleKyError(err)
    }
  }

  getEvent = async (eventId: number) => {
    try {
      const resp = await this.client
        .get(`${API_PATH.EVENT_GET}/${eventId}`)
        .json()
      return EventGetResp.parse(resp).event
    } catch (err) {
      return this.handleKyError(err)
    }
  }

  saveEvent = async (payload: SaveEventParams) => {
    try {
      const csrfToken = this.getCsrfToken()
      const resp = await this.client
        .post(API_PATH.EVENT_SAVE, {
          json: payload,
          headers: { 'X-CSRF-Token': csrfToken },
        })
        .json()
      return EventSaveResp.parse(resp)
    } catch (err) {
      return this.handleKyError(err)
    }
  }
}

/** Single API client to be used from client-side calls.
 *  When using SSR, you should construct a new ApiClient
 *  using the cookies.
 */
export const apiClient = new ApiClient()
