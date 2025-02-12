import { API_PATH, ApiClient } from 'lib/api'
import { QueryClient } from '@tanstack/react-query'

export const fetchSession = async (cookies?: string) => {
  const queryClient = new QueryClient()
  const apiClient = new ApiClient(cookies)

  const data = await queryClient.fetchQuery({
    queryKey: [API_PATH.USER_ME],
    queryFn: apiClient.getAuthedUser,
  })

  return {
    user: data?.user
      ? {
          ...data.user,
          role: data.mainRole,
        }
      : null,
  }
}
