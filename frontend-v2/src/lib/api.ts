import ky, { HTTPError, KyInstance } from 'ky'
import { z } from 'zod'
import { QueryActivistOptions, QueryActivistResult } from './api/activists'

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
  EVENT_LIST: 'event/list',
  EVENT_DELETE: 'event/delete',
}

export const StaticResourcesHashResp = z.object({
  hash: z.string(),
})

export const Role = z.enum([
  'admin',
  'organizer',
  'attendance',
  'intl_coordinator',
])
export type Role = z.infer<typeof Role>

export const AuthedUserSchema = z.object({
  ChapterID: z.number(),
  ChapterName: z.string(),
  Disabled: z.boolean(),
  Email: z.string(),
  ID: z.number(),
  Name: z.string(),
  Roles: z.array(Role),
})
export type AuthedUser = z.infer<typeof AuthedUserSchema>

const AuthedUserResp = z.object({
  user: AuthedUserSchema,
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

const ChapterOrganizerSchema = z.object({
  Name: z.string(),
  Email: z.string().optional(),
  Phone: z.string().optional(),
  Facebook: z.string().optional(),
  Twitter: z.string().optional(),
  Instagram: z.string().optional(),
  Linkedin: z.string().optional(),
})
export type ChapterOrganizer = z.infer<typeof ChapterOrganizerSchema>

const ChapterWithOrganizersSchema = z.object({
  ChapterID: z.number(),
  Name: z.string(),
  Organizers: z
    .array(ChapterOrganizerSchema)
    .nullable()
    .transform((v) => v ?? []),
})

const ChapterListWithOrganizersResp = z.object({
  chapters: z.array(ChapterWithOrganizersSchema),
})

export const CHAPTER_ORGANIZERS_QUERY_KEY = [
  API_PATH.CHAPTER_LIST,
  'withOrganizers',
] as const

export interface InternationalOrganizer {
  chapterName: string
  chapterId: number
  name: string
  email: string
  phone: string
  facebook: string
  twitter: string
  instagram: string
  linkedin: string
}

export function flattenChapterOrganizers(
  chapters: z.infer<typeof ChapterListWithOrganizersResp>['chapters'],
): InternationalOrganizer[] {
  return chapters.flatMap((chapter) =>
    chapter.Organizers.map((org) => ({
      chapterName: chapter.Name,
      chapterId: chapter.ChapterID,
      name: org.Name,
      email: org.Email ?? '',
      phone: org.Phone ?? '',
      facebook: org.Facebook ?? '',
      twitter: org.Twitter ?? '',
      instagram: org.Instagram ?? '',
      linkedin: org.Linkedin ?? '',
    })),
  )
}

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
    z
      .object({
        id: z.number(),
        name: z.string(),
        email: z.boolean(),
        phone: z.boolean(),
        last_updated: z.number(),
        last_event_date: z.number(),
      })
      .transform((a) => ({
        id: a.id,
        name: a.name,
        email: a.email,
        phone: a.phone,
        lastUpdated: a.last_updated * 1000,
        lastEventDate: a.last_event_date * 1000,
      })),
  ),
  hidden_ids: z.array(z.number()),
})

export type ActivistListBasic = z.infer<typeof ActivistListBasicResp>

// Re-export activist search types from dedicated module
export {
  ActivistJSON,
  ActivistColumnName,
  QueryActivistOptions,
  QueryActivistResult,
} from './api/activists'
export type {
  ApiDateRangeFilter,
  ApiIntRangeFilter,
  ActivistJSON as ActivistJSONType,
  ActivistColumnName as ActivistColumnNameType,
  QueryActivistOptions as QueryActivistOptionsType,
  QueryActivistResult as QueryActivistResultType,
} from './api/activists'

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

const EventListItemSchema = z.object({
  event_id: z.number(),
  event_name: z.string(),
  event_date: z.string(),
  event_type: z.string(),
  attendees: z
    .array(z.string())
    .nullable()
    .transform((v) => v ?? []),
  attendee_emails: z
    .array(z.string())
    .nullable()
    .transform((v) => v ?? []),
})
export type EventListItem = z.infer<typeof EventListItemSchema>

const EventListResp = z.array(EventListItemSchema)

export const EVENT_TYPE_VALUES = [
  'noConnections',
  'Connection',
  'Action',
  'Campaign Action',
  'Community',
  'Frontline Surveillance',
  'Meeting',
  'Outreach',
  'Animal Care',
  'Training',
  'mpiDA',
  'mpiCOM',
] as const

export type EventType = (typeof EVENT_TYPE_VALUES)[number]

export interface EventListParams {
  event_name?: string
  event_activist?: string
  event_date_start: string
  event_date_end: string
  event_type: EventType
}

const EventDeleteResp = z.object({
  status: z.literal('success'),
})

const ApiErrorResp = z.object({
  status: z.literal('error'),
  message: z.string(),
})

export class HTTPStatusError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message)
    this.name = 'HTTPStatusError'
  }
}

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
        throw new HTTPStatusError(err.response.status, parsed.data.message)
      }
      throw new HTTPStatusError(err.response.status, err.message)
    }
    throw err
  }

  getAuthedUser = async (signal?: AbortSignal) => {
    try {
      const resp = await this.client.get(API_PATH.USER_ME, { signal }).json()
      return AuthedUserResp.parse(resp)
    } catch (err) {
      return this.handleKyError(err)
    }
  }

  getStaticResourceHash = async (signal?: AbortSignal) => {
    const resp = await this.client
      .get(API_PATH.STATIC_RESOURCE_HASH, { signal })
      .json()
    return StaticResourcesHashResp.parse(resp)
  }

  async fetchCsrfToken(signal?: AbortSignal) {
    try {
      const resp = await this.client.get(API_PATH.CSRF_TOKEN, { signal }).json()
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

  getActivistNames = async (signal?: AbortSignal) => {
    const resp = await this.client
      .get(API_PATH.ACTIVIST_NAMES_GET, { signal })
      .json()
    return ActivistNamesResp.parse(resp)
  }

  getActivistListBasic = async (
    modifiedSince?: string,
    signal?: AbortSignal,
  ) => {
    const options = modifiedSince
      ? { searchParams: { modified_since: modifiedSince }, signal }
      : { signal }

    const resp = await this.client
      .get(API_PATH.ACTIVIST_LIST_BASIC, options)
      .json()
    return ActivistListBasicResp.parse(resp)
  }

  searchActivists = async (
    options: QueryActivistOptions,
    signal?: AbortSignal,
  ) => {
    try {
      const resp = await this.client
        .post(API_PATH.ACTIVISTS_SEARCH, { json: options, signal })
        .json()
      const parsed = QueryActivistResult.parse(resp)
      return fillActivistBlankNumericFieldsWithZero(parsed, options.columns)
    } catch (err) {
      return this.handleKyError(err)
    }
  }

  getChapterList = async (signal?: AbortSignal) => {
    try {
      const resp = await this.client
        .get(API_PATH.CHAPTER_LIST, { signal })
        .json()
      return ChapterListResp.parse(resp).chapters
    } catch (err) {
      return this.handleKyError(err)
    }
  }

  getChapterListWithOrganizers = async (signal?: AbortSignal) => {
    try {
      const resp = await this.client
        .get(API_PATH.CHAPTER_LIST, { signal })
        .json()
      return ChapterListWithOrganizersResp.parse(resp).chapters
    } catch (err) {
      return this.handleKyError(err)
    }
  }

  getUsers = async (signal?: AbortSignal) => {
    try {
      const resp = await this.client.get(API_PATH.USERS, { signal }).json()
      return UserListResp.parse(resp).users
    } catch (err) {
      return this.handleKyError(err)
    }
  }

  getUser = async (userId: number, signal?: AbortSignal) => {
    try {
      const resp = await this.client
        .get(`${API_PATH.USERS}/${userId}`, { signal })
        .json()
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

  getEvent = async (eventId: number, signal?: AbortSignal) => {
    try {
      const resp = await this.client
        .get(`${API_PATH.EVENT_GET}/${eventId}`, { signal })
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

  getEventList = async (params: EventListParams, signal?: AbortSignal) => {
    try {
      const body = new URLSearchParams({
        ...(params.event_name && { event_name: params.event_name }),
        ...(params.event_activist && { event_activist: params.event_activist }),
        event_date_start: params.event_date_start,
        event_date_end: params.event_date_end,
        event_type: params.event_type,
      })
      const resp = await this.client
        .post(API_PATH.EVENT_LIST, { body, signal })
        .json()
      return EventListResp.parse(resp)
    } catch (err) {
      return this.handleKyError(err)
    }
  }

  deleteEvent = async (eventId: number) => {
    try {
      const csrfToken = this.getCsrfToken()
      const body = new URLSearchParams({ event_id: String(eventId) })
      const resp = await this.client
        .post(API_PATH.EVENT_DELETE, {
          body,
          headers: { 'X-CSRF-Token': csrfToken },
        })
        .json()
      return EventDeleteResp.parse(resp)
    } catch (err) {
      return this.handleKyError(err)
    }
  }
}

const BLANK_TO_ZERO_FIELDS = [
  'total_events',
  'months_since_last_action',
  'total_points',
  'total_interactions',
] as const

// If a column is requested but not set, this is due to use of "omitempty" in Go JSON serialization. The real value of
// the field is still 0, so this fills in those missing values until we implement a more semantic serialization in Go.
function fillActivistBlankNumericFieldsWithZero(
  result: z.infer<typeof QueryActivistResult>,
  requestedColumns: QueryActivistOptions['columns'],
): z.infer<typeof QueryActivistResult> {
  const requested = new Set(requestedColumns)
  const fillableFields = BLANK_TO_ZERO_FIELDS.filter((field) =>
    requested.has(field),
  )
  if (fillableFields.length === 0) {
    return result
  }

  return {
    ...result,
    activists: result.activists.map((activist) => {
      const normalized = { ...activist }
      for (const field of fillableFields) {
        if (normalized[field] === undefined) {
          normalized[field] = 0
        }
      }
      return normalized
    }),
  }
}

/** Single API client to be used from client-side calls.
 *  When using SSR, you should construct a new ApiClient
 *  using the cookies.
 */
export const apiClient = new ApiClient()
