import { cache } from 'react'
import { API_PATH, ApiClient, type AuthedUser } from '@/lib/api'
import { QueryClient } from '@tanstack/react-query'
import { getCookies } from '@/lib/auth'

export const getCachedSession = cache(async () =>
  fetchSession(await getCookies()),
)

export const fetchSession = async (cookies?: string) => {
  const queryClient = new QueryClient()
  const apiClient = new ApiClient(cookies)

  const data = await queryClient.fetchQuery({
    queryKey: [API_PATH.USER_ME],
    queryFn: ({ signal }) => apiClient.getAuthedUser(signal),
  })

  return {
    user: data?.user ?? null,
  }
}
