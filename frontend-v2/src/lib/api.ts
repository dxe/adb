import ky from 'ky'
import { z } from 'zod'

export const getBaseUrl = () => {
  return typeof window === 'undefined'
    ? process.env.NEXT_PUBLIC_API_BASE_URL
    : '/'
}

const apiClient = ky.extend({
  prefixUrl: getBaseUrl(),
})

export const API_PATH = {
  STATIC_RESOURCE_HASH: 'static_resources_hash',
  ACTIVIST_NAMES_GET: 'activist_names/get',
  USER_ME: 'user/me',
}

export const StaticResourcesHashResp = z.object({
  hash: z.string(),
})

export const getStaticResourceHash = async () => {
  const resp = await apiClient.get(API_PATH.STATIC_RESOURCE_HASH).json()
  return StaticResourcesHashResp.parse(resp)
}

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

export const getAuthedUser = async () => {
  try {
    const resp = await apiClient.get(API_PATH.USER_ME).json()
    return AuthedUserResp.parse(resp)
  } catch (err) {
    console.error(`Error fetching authed user: ${err}`)
    return {
      user: null,
      mainRole: null,
    }
  }
}

export const ActivistNamesResp = z.object({
  activist_names: z.array(z.string()),
})

export const getActivistNames = async () => {
  const resp = await apiClient.get(API_PATH.ACTIVIST_NAMES_GET).json()
  return ActivistNamesResp.parse(resp)
}
