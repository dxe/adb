import { API_PATH, getAuthedUser } from '@/lib/api'
import { useQuery } from '@tanstack/react-query'

export const useSession = () => {
  const { data, isLoading } = useQuery({
    queryKey: [API_PATH.USER_ME],
    queryFn: getAuthedUser,
  })

  return {
    user: {
      ...data?.user,
      role: data?.mainRole,
    },
    isLoading: isLoading,
  }
}
