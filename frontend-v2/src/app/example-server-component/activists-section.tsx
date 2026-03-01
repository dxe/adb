import {
  dehydrate,
  HydrationBoundary,
  QueryClient,
} from '@tanstack/react-query'
import Activists from './activists'
import { API_PATH, ApiClient } from '@/lib/api'
import { getCookies } from '@/lib/auth'

export async function ActivistsSection() {
  const apiClient = new ApiClient(await getCookies())
  const queryClient = new QueryClient()

  await queryClient.prefetchQuery({
    queryKey: [API_PATH.ACTIVIST_NAMES_GET],
    queryFn: apiClient.getActivistNames,
  })

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <Activists />
    </HydrationBoundary>
  )
}
