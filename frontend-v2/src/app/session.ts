import { cache } from 'react'
import { API_PATH, ApiClient, HTTPStatusError } from '@/lib/api'
import { QueryClient } from '@tanstack/react-query'
import { getCookies } from '@/lib/auth'

export const getCachedSession = cache(async () =>
  fetchSession(await getCookies()),
)

export const fetchSession = async (cookies?: string) => {
  const queryClient = new QueryClient()
  const apiClient = new ApiClient(cookies)

  try {
    const data = await queryClient.fetchQuery({
      queryKey: [API_PATH.USER_ME],
      queryFn: ({ signal }) => apiClient.getAuthedUser(signal),
    })

    return {
      user: data.user,
    }
  } catch (err) {
    if (
      err instanceof HTTPStatusError &&
      // Backend uses 400 for unauthenticated and 403 for other issues, such as
      // the user not having any roles.
      (err.status === 400 || err.status === 403)
    ) {
      return {
        user: null,
      }
    }
    throw err
  }
}
