import { API_PATH, apiClient } from '@/lib/api'
import { useQuery } from '@tanstack/react-query'

export const useSession = () => {
  const { data, isLoading } = useQuery({
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
    isLoading: isLoading,
  }
}
