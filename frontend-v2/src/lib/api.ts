import ky, { HTTPError, KyInstance } from 'ky'
import { z } from 'zod'

export const API_PATH = {
  STATIC_RESOURCE_HASH: 'static_resources_hash',
  ACTIVIST_NAMES_GET: 'activist_names/get',
  USER_ME: 'user/me',
  CSRF_TOKEN: 'api/csrf-token',
  CHAPTER_LIST: 'chapter/list',
}

export const StaticResourcesHashResp = z.object({
  hash: z.string(),
})

const Role = z.enum(['admin', 'organizer', 'attendance', 'non-sfbay'])

const AuthedUserResp = z.object({
  user: z.object({
    Admin: z.boolean(),
    ChapterID: z.number(),
    ChapterName: z.string(),
    Disabled: z.boolean(),
    Email: z.string(),
    ID: z.number(),
    Name: z.string(),
    Roles: z
      .array(z.object({ Role: Role }))
      .transform((roles) => roles.map((it) => it.Role)),
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

export const ActivistNamesResp = z.object({
  activist_names: z.array(z.string()),
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

  getChapterList = async () => {
    const resp = await this.client.get(API_PATH.CHAPTER_LIST).json()
    return ChapterListResp.parse(resp).chapters
  }
}

/** Single API client to be used from client-side calls.
 *  When using SSR, you should construct a new ApiClient
 *  using the cookies.
 */
export const apiClient = new ApiClient()
