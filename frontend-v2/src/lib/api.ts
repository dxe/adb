import ky, { HTTPError, KyInstance } from 'ky'
import { z } from 'zod'
import {
  ActivistJSON,
  ActivistPatchInput,
  QueryActivistCountOptions,
  QueryActivistCountResult,
  QueryActivistOptions,
  QueryActivistResult,
} from './api/activists'
import {
  BLANK_TO_FALSE_FIELDS,
  BLANK_TO_ZERO_FIELDS,
} from '@/app/(authed)/activists/column-definitions'

export const API_PATH = {
  STATIC_RESOURCE_HASH: 'static_resources_hash',
  ACTIVIST_NAMES_GET: 'activist_names/get',
  ACTIVIST_LIST_BASIC: 'activist/list_basic',
  ACTIVISTS_SEARCH: 'api/activists',
  ACTIVISTS_COUNT: 'api/activists/count',
  ACTIVISTS_EXPORT: 'api/activists/export',
  ACTIVISTS_EXPORT_SPOKE: 'api/activists/export/spoke',
  ACTIVISTS_DEBUG_QUERY: 'api/activists/debug-query',
  ACTIVIST_GET: 'api/activists',
  // The real URL puts the activist id in the middle
  // (api/activists/{id}/timeline); this entry exists to key the query cache.
  // TODO: avoid conflating query keys and real API paths.
  ACTIVIST_TIMELINE: 'api/activists/timeline',
  ACTIVIST_HIDE: 'activist/hide',
  ACTIVIST_MERGE: 'activist/merge',
  INTERACTION_SAVE: 'interaction/save',
  USER_ME: 'user/me',
  CSRF_TOKEN: 'api/csrf-token',
  CHAPTER_LIST: 'chapter/list',
  USERS: 'api/users',
  USERS_ASSIGNABLE: 'api/users/assignable',
  EVENT_GET: 'event/get',
  EVENT_SAVE: 'event/save',
  EVENT_LIST: 'event/list',
  EVENT_DELETE: 'event/delete',
  COACHING_SAVE: 'connection/save',
  ADMIN_SEND_TEST_EMAIL: 'api/admin/send-test-email',
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
  // Referrer-restricted Google Places key, served from the backend config so
  // it works in all environments without a build-time env var. Defaults to ''
  // (not null/undefined) so consumers keep a plain `string` and treat the
  // not-configured case as a simple falsy check rather than a null guard.
  googlePlacesApiKey: z.string().optional().default(''),
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
export type Chapter = z.infer<typeof ChapterListResp>['chapters'][number]

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

const AssignableUserSchema = z.object({
  id: z.number(),
  name: z.string(),
})
export type AssignableUser = z.infer<typeof AssignableUserSchema>

const AssignableUserListResp = z.object({
  users: z.array(AssignableUserSchema),
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

// One entry in an activist's engagement timeline: either an event they
// attended or an interaction an organizer logged with them. `date` is the
// calendar date (YYYY-MM-DD) the server filed the item under and is what both
// kinds are ordered by; interactions additionally carry the exact instant.
const TimelineEventPayloadSchema = z.object({
  name: z.string(),
  type: z.string(),
})

const TimelineInteractionPayloadSchema = z.object({
  method: z.string(),
  outcome: z.string(),
  notes: z.string(),
  user_name: z.string(),
})

// Fields every timeline item carries, whatever its variant. `date` is the day
// the server filed the item under and is what the ordering agrees with, so it
// is rendered as sent rather than re-derived from `timestamp`. `has_time` is
// false when the server had no real time of day and supplied a placeholder
// instant just to sort the item (an event with no start time), so `timestamp`
// is only safe to show as a clock time when it is true.
const TimelineItemEnvelope = {
  id: z.number(),
  date: z.string(),
  timestamp: z.string(),
  has_time: z.boolean(),
}

const TimelineEventItemSchema = z.object({
  ...TimelineItemEnvelope,
  type: z.literal('event'),
  event: TimelineEventPayloadSchema,
})

const TimelineInteractionItemSchema = z.object({
  ...TimelineItemEnvelope,
  type: z.literal('interaction'),
  interaction: TimelineInteractionPayloadSchema,
})

export const ActivistTimelineItemSchema = z.discriminatedUnion('type', [
  TimelineEventItemSchema,
  TimelineInteractionItemSchema,
])
export type ActivistTimelineItem = z.infer<typeof ActivistTimelineItemSchema>
export type ActivistTimelineEventItem = z.infer<typeof TimelineEventItemSchema>
export type ActivistTimelineInteractionItem = z.infer<
  typeof TimelineInteractionItemSchema
>
export type ActivistTimelineEventPayload = z.infer<
  typeof TimelineEventPayloadSchema
>
export type ActivistTimelineInteractionPayload = z.infer<
  typeof TimelineInteractionPayloadSchema
>

export const ActivistTimelineResp = z.object({
  items: z.array(ActivistTimelineItemSchema),
  // The server caps how many items it returns; true means older items exist
  // that were not sent.
  truncated: z.boolean(),
})
export type ActivistTimeline = z.infer<typeof ActivistTimelineResp>

// Payload accepted by POST /interaction/save when logging a new interaction.
// Mirrors model.Interaction in server/src/model/interactions.go — keep in sync.
// `id` is omitted so the server treats this as an insert, and `user_id` is
// omitted so it attributes the interaction to the authed user.
//
// The three follow-up fields act on the activist's followup_date rather than on
// the interaction row: reset_followup clears it, and set_followup moves it
// followup_days into the future. The server ignores followup_days unless
// set_followup is true, and the two flags are mutually exclusive in the UI.
export const SaveInteractionInput = z.object({
  activist_id: z.number(),
  method: z.string(),
  outcome: z.string(),
  notes: z.string(),
  assign_self: z.boolean(),
  reset_followup: z.boolean(),
  set_followup: z.boolean(),
  followup_days: z.number(),
})
export type SaveInteractionInput = z.infer<typeof SaveInteractionInput>

// Re-export activist search types from dedicated module
export {
  ActivistJSON,
  ActivistColumnName,
  ActivistPatchInput,
  QueryActivistOptions,
  QueryActivistShape,
  QueryActivistResult,
  QueryActivistCountOptions,
} from './api/activists'
export type {
  ApiDateRangeFilter,
  ApiIntRangeFilter,
  ActivistJSON as ActivistJSONType,
  ActivistColumnName as ActivistColumnNameType,
  ActivistPatchInput as ActivistPatchInputType,
  QueryActivistOptions as QueryActivistOptionsType,
  QueryActivistShape as QueryActivistShapeType,
  QueryActivistResult as QueryActivistResultType,
  QueryActivistCountOptions as QueryActivistCountOptionsType,
} from './api/activists'

const ActivistGetResp = z.object({
  activist: ActivistJSON,
})

// The location attached to an event: a free-text display name plus optional geo
// data (a Google Place id and/or coordinates). Grouped under `location` on the
// API.
const EventLocationSchema = z.object({
  google_place_id: z.string().nullish(),
  name: z.string().nullish(),
  formatted_address: z.string().nullish(),
  lat: z.number().nullish(),
  lng: z.number().nullish(),
})

const EventGetResp = z.object({
  event: z.object({
    event_name: z.string(),
    event_type: z.string(),
    event_date: z.string(),
    attendees: z.array(z.string()).nullable(),
    suppress_survey: z.boolean(),
    // Advance-event fields (optional; absent on legacy attendance events).
    is_online: z.boolean().nullish(),
    description: z.string().nullish(),
    start_time: z.string().nullish(),
    end_time: z.string().nullish(),
    timezone: z.string().nullish(),
    is_public: z.boolean().nullish(),
    // Location: a free-text name plus optional geo data. Null for
    // attendance/online events.
    location: EventLocationSchema.nullish(),
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
  // Advance-event fields (optional).
  is_online?: boolean
  description?: string
  start_time?: string
  end_time?: string
  timezone?: string
  is_public?: boolean
  // Location: a free-text name plus optional geo data. Only the name is
  // required; the geo fields are omitted when empty (e.g. a manual lat/lng
  // location with no Google place or formatted address).
  location?: {
    google_place_id?: string
    name: string
    formatted_address?: string
    lat?: number
    lng?: number
  }
}

const EventSaveResp = z.object({
  status: z.literal('success'),
  event_id: z.number(),
  attendees: z.array(z.string()).nullish(),
})

const EventListItemSchema = z.object({
  event_id: z.number(),
  event_name: z.string(),
  event_date: z.string(),
  event_type: z.string(),
  start_time: z.string().nullish(),
  end_time: z.string().nullish(),
  timezone: z.string().nullish(),
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

const SuccessResp = z.object({
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

// Module-level cache for the CSRF token, shared across ApiClient instances.
let _csrfTokenCache: string | undefined
let _csrfTokenPending: Promise<string | undefined> | null = null

async function getCsrfToken(client: ApiClient): Promise<string | undefined> {
  if (typeof window === 'undefined') return undefined
  if (_csrfTokenCache !== undefined) return _csrfTokenCache
  if (!_csrfTokenPending) {
    _csrfTokenPending = client.fetchCsrfToken().then(
      (token) => {
        _csrfTokenCache = token
        _csrfTokenPending = null
        return token
      },
      () => {
        _csrfTokenPending = null
        return undefined
      },
    )
  }
  return _csrfTokenPending
}

// Drops the cached token so the next getCsrfToken() refetches from the server.
function invalidateCsrfToken() {
  _csrfTokenCache = undefined
  _csrfTokenPending = null
}

// Whether the given error is a 403 caused by a rejected CSRF token.
// Distinguishes from authorization 403.
function isCsrfRejection(err: unknown): boolean {
  return (
    err instanceof HTTPError &&
    err.response.status === 403 &&
    typeof err.data === 'string' &&
    // The CSRF middleware's body is "Forbidden - <reason>" and its token
    // reasons contain "CSRF".
    err.data.includes('CSRF')
  )
}

export function preloadCsrfToken() {
  void getCsrfToken(new ApiClient())
}

export class ApiClient {
  private client: KyInstance

  constructor(cookies?: string) {
    this.client = ky.extend({
      prefix:
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

  private handleKyError(err: unknown): never {
    if (err instanceof HTTPError) {
      // ky consumes the response body when it builds the error and exposes the
      // pre-parsed payload on err.data.
      const parsed = ApiErrorResp.safeParse(err.data)
      if (parsed.success) {
        throw new HTTPStatusError(err.response.status, parsed.data.message)
      }
      throw new HTTPStatusError(err.response.status, err.message)
    }
    throw err
  }

  // Some legacy Go handlers report errors via a 200 response with body
  // {status: "error", message: "..."}, so the ky `HTTPError` path never fires.
  // Call this on the parsed JSON before validating it as a success response.
  private throwIfApiError(resp: unknown): void {
    const apiError = ApiErrorResp.safeParse(resp)
    if (apiError.success) {
      throw new HTTPStatusError(200, apiError.data.message)
    }
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

  private getCsrfToken(): Promise<string | undefined> {
    return getCsrfToken(this)
  }

  // Runs a CSRF-protected request, passing `run` the current token to attach as
  // the X-CSRF-Token header.
  //
  // If the token has gone stale and the request is rejected for CSRF, refetches
  // a fresh token and retries once. Other errors propagate untouched.
  private async withCsrf<T>(
    run: (csrfToken: string | undefined) => Promise<T>,
  ): Promise<T> {
    try {
      return await run(await this.getCsrfToken())
    } catch (err) {
      if (isCsrfRejection(err)) {
        invalidateCsrfToken()
        return await run(await this.getCsrfToken())
      }
      throw err
    }
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
      const result = QueryActivistResult.parse(resp)
      fillBlankFieldsInQueryActivistResult(result, options.shape.columns)
      return result
    } catch (err) {
      return this.handleKyError(err)
    }
  }

  countActivists = async (
    options: QueryActivistCountOptions,
    signal?: AbortSignal,
  ): Promise<QueryActivistCountResult> => {
    try {
      const resp = await this.client
        .post(API_PATH.ACTIVISTS_COUNT, { json: options, signal })
        .json()
      return QueryActivistCountResult.parse(resp)
    } catch (err) {
      return this.handleKyError(err)
    }
  }

  exportActivistsCsv = async (
    options: QueryActivistOptions,
    signal?: AbortSignal,
  ) => {
    try {
      const resp = await this.client.post(API_PATH.ACTIVISTS_EXPORT, {
        json: options,
        signal,
        timeout: false,
      })
      return await resp.blob()
    } catch (err) {
      return this.handleKyError(err)
    }
  }

  // Exports the current columns as tab-separated text, for copying to the
  // clipboard so it pastes into spreadsheet cells.
  exportActivistsTsvText = async (
    options: QueryActivistOptions,
    signal?: AbortSignal,
  ) => {
    try {
      const resp = await this.client.post(API_PATH.ACTIVISTS_EXPORT, {
        searchParams: { format: 'tsv' },
        json: options,
        signal,
        timeout: false,
      })
      return await resp.text()
    } catch (err) {
      return this.handleKyError(err)
    }
  }

  // Exports activists in the Spoke dialer layout. The server hard-codes the
  // CSV columns; options.shape.columns must be empty.
  exportActivistsSpokeCsv = async (
    options: QueryActivistOptions,
    signal?: AbortSignal,
  ) => {
    try {
      const resp = await this.client.post(API_PATH.ACTIVISTS_EXPORT_SPOKE, {
        json: options,
        signal,
        timeout: false,
      })
      return await resp.blob()
    } catch (err) {
      return this.handleKyError(err)
    }
  }

  debugQueryActivists = async (
    options: QueryActivistOptions,
    signal?: AbortSignal,
  ): Promise<{ id: number }> => {
    try {
      const resp = await this.client
        .post(API_PATH.ACTIVISTS_DEBUG_QUERY, { json: options, signal })
        .json()
      return z.object({ id: z.number() }).parse(resp)
    } catch (err) {
      return this.handleKyError(err)
    }
  }

  getActivist = async (activistId: number, signal?: AbortSignal) => {
    try {
      const resp = await this.client
        .get(`${API_PATH.ACTIVIST_GET}/${activistId}`, { signal })
        .json()
      const activist = ActivistGetResp.parse(resp).activist
      fillActivistBlankFields(activist)
      return activist
    } catch (err) {
      return this.handleKyError(err)
    }
  }

  patchActivist = async (
    activistId: number,
    patch: ActivistPatchInput,
    signal?: AbortSignal,
  ) => {
    try {
      const activist = await this.withCsrf(async (csrfToken) => {
        const resp = await this.client
          .patch(`${API_PATH.ACTIVIST_GET}/${activistId}`, {
            json: patch,
            headers: { 'X-CSRF-Token': csrfToken },
            signal,
          })
          .json()
        return ActivistGetResp.parse(resp).activist
      })
      fillActivistBlankFields(activist)
      return activist
    } catch (err) {
      return this.handleKyError(err)
    }
  }

  getActivistTimeline = async (activistId: number, signal?: AbortSignal) => {
    try {
      const resp = await this.client
        .get(`${API_PATH.ACTIVIST_GET}/${activistId}/timeline`, { signal })
        .json()
      return ActivistTimelineResp.parse(resp)
    } catch (err) {
      return this.handleKyError(err)
    }
  }

  hideActivist = async (activistId: number, signal?: AbortSignal) => {
    try {
      // TODO: pass X-CSRF-Token header once the backend requires it for this endpoint.
      const resp = await this.client
        .post(API_PATH.ACTIVIST_HIDE, {
          json: { id: activistId },
          signal,
        })
        .json()
      this.throwIfApiError(resp)
      return SuccessResp.parse(resp)
    } catch (err) {
      return this.handleKyError(err)
    }
  }

  mergeActivist = async (
    currentActivistId: number,
    targetActivistName: string,
    signal?: AbortSignal,
  ) => {
    try {
      // TODO: pass X-CSRF-Token header once the backend requires it for this endpoint.
      const resp = await this.client
        .post(API_PATH.ACTIVIST_MERGE, {
          json: {
            current_activist_id: currentActivistId,
            target_activist_name: targetActivistName,
          },
          signal,
        })
        .json()
      this.throwIfApiError(resp)
      return SuccessResp.parse(resp)
    } catch (err) {
      return this.handleKyError(err)
    }
  }

  // The response also carries the updated activist, but callers refetch the
  // activist and timeline queries instead of trusting a second shape here.
  saveInteraction = async (payload: SaveInteractionInput) => {
    try {
      return await this.withCsrf(async (csrfToken) => {
        const resp = await this.client
          .post(API_PATH.INTERACTION_SAVE, {
            json: payload,
            headers: { 'X-CSRF-Token': csrfToken },
          })
          .json()
        this.throwIfApiError(resp)
        return SuccessResp.parse(resp)
      })
    } catch (err) {
      return this.handleKyError(err)
    }
  }

  getChapterList = async (signal?: AbortSignal) => {
    try {
      const resp = await this.client
        .get(API_PATH.CHAPTER_LIST, { signal })
        .json()
      this.throwIfApiError(resp)
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
      this.throwIfApiError(resp)
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

  getAssignableUsers = async (signal?: AbortSignal) => {
    try {
      const resp = await this.client
        .get(API_PATH.USERS_ASSIGNABLE, { signal })
        .json()
      return AssignableUserListResp.parse(resp).users
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
      return await this.withCsrf(async (csrfToken) => {
        const resp = await this.client
          .post(API_PATH.USERS, {
            json: payload,
            headers: { 'X-CSRF-Token': csrfToken },
          })
          .json()
        return UserSaveResp.parse(resp).user
      })
    } catch (err) {
      return this.handleKyError(err)
    }
  }

  updateUser = async (payload: User) => {
    try {
      return await this.withCsrf(async (csrfToken) => {
        const resp = await this.client
          .put(`${API_PATH.USERS}/${payload.id}`, {
            json: payload,
            headers: { 'X-CSRF-Token': csrfToken },
          })
          .json()
        return UserSaveResp.parse(resp).user
      })
    } catch (err) {
      return this.handleKyError(err)
    }
  }

  sendTestEmail = async (email: string) => {
    try {
      return await this.withCsrf(async (csrfToken) => {
        const resp = await this.client
          .post(API_PATH.ADMIN_SEND_TEST_EMAIL, {
            json: { email },
            headers: { 'X-CSRF-Token': csrfToken },
          })
          .json()
        this.throwIfApiError(resp)
        return SuccessResp.parse(resp)
      })
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
      return await this.withCsrf(async (csrfToken) => {
        const resp = await this.client
          .post(API_PATH.EVENT_SAVE, {
            json: payload,
            headers: { 'X-CSRF-Token': csrfToken },
          })
          .json()
        this.throwIfApiError(resp)
        return EventSaveResp.parse(resp)
      })
    } catch (err) {
      return this.handleKyError(err)
    }
  }

  saveCoaching = async (payload: SaveEventParams) => {
    try {
      return await this.withCsrf(async (csrfToken) => {
        const resp = await this.client
          .post(API_PATH.COACHING_SAVE, {
            json: payload,
            headers: { 'X-CSRF-Token': csrfToken },
          })
          .json()
        this.throwIfApiError(resp)
        return EventSaveResp.parse(resp)
      })
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
      this.throwIfApiError(resp)
      return EventListResp.parse(resp)
    } catch (err) {
      return this.handleKyError(err)
    }
  }

  deleteEvent = async (eventId: number) => {
    try {
      const body = new URLSearchParams({ event_id: String(eventId) })
      return await this.withCsrf(async (csrfToken) => {
        const resp = await this.client
          .post(API_PATH.EVENT_DELETE, {
            body,
            headers: { 'X-CSRF-Token': csrfToken },
          })
          .json()
        this.throwIfApiError(resp)
        return SuccessResp.parse(resp)
      })
    } catch (err) {
      return this.handleKyError(err)
    }
  }
}

// If a column is requested but the field on the activist is not set
// (undefined), this is due to use of "omitempty" in Go JSON serialization.
// The real value of the field is still 0, so these fields need to have their
// zero values added back until we implement smarter serialization in Go.
type BlankToZeroField = (typeof BLANK_TO_ZERO_FIELDS)[number]
type BlankToFalseField = (typeof BLANK_TO_FALSE_FIELDS)[number]

function fillActivistBlankFields(
  activist: ActivistJSON,
  fields: {
    zero: ReadonlyArray<BlankToZeroField>
    false: ReadonlyArray<BlankToFalseField>
  } = {
    zero: BLANK_TO_ZERO_FIELDS,
    false: BLANK_TO_FALSE_FIELDS,
  },
) {
  const numericFields = activist as ActivistJSON &
    Record<BlankToZeroField, number | undefined>
  for (const field of fields.zero) {
    if (numericFields[field] === undefined) {
      numericFields[field] = 0
    }
  }
  const booleanFields = activist as ActivistJSON &
    Record<BlankToFalseField, boolean | undefined>
  for (const field of fields.false) {
    if (booleanFields[field] === undefined) {
      booleanFields[field] = false
    }
  }
}

function fillBlankFieldsInQueryActivistResult(
  result: z.infer<typeof QueryActivistResult>,
  requestedColumns: string[],
): void {
  const columns = new Set(requestedColumns)
  const blankNumericFields = BLANK_TO_ZERO_FIELDS.filter((field) =>
    columns.has(field),
  )
  const blankBooleanFields = BLANK_TO_FALSE_FIELDS.filter((field) =>
    columns.has(field),
  )
  if (blankNumericFields.length === 0 && blankBooleanFields.length === 0) {
    return
  }

  result.activists.forEach((activist: ActivistJSON) => {
    fillActivistBlankFields(activist, {
      zero: blankNumericFields,
      false: blankBooleanFields,
    })
  })
}

/** Single API client to be used from client-side calls.
 *  When using SSR, you should construct a new ApiClient
 *  using the cookies.
 */
export const apiClient = new ApiClient()
