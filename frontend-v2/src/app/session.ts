import { API_PATH, apiClient } from 'lib/api'
import { QueryClient } from '@tanstack/react-query'

export const fetchSession = async () => {
  const queryClient = new QueryClient()

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
