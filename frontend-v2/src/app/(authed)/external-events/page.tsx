import { forbidden } from 'next/navigation'
import {
  dehydrate,
  HydrationBoundary,
  QueryClient,
} from '@tanstack/react-query'
import { ContentWrapper } from '@/app/content-wrapper'
import { getCachedSession } from '@/app/session'
import { API_PATH, ApiClient } from '@/lib/api'
import { getCookies } from '@/lib/auth'
import ExternalEventsPage from './external-events-page'

export default async function ExternalEventsListPage() {
  const session = await getCachedSession()
  if (!session.user?.Roles.includes('admin')) {
    forbidden()
  }

  const apiClient = new ApiClient(await getCookies())
  const queryClient = new QueryClient()

  // Intentionally use prefetchQuery here: this endpoint isn't itself
  // admin-gated (the legacy Vue page calls it unauthenticated), so there's no
  // 403/404 to redirect on and a failed prefetch can just fall through to the
  // client-side useQuery below.
  await queryClient.prefetchQuery({
    queryKey: [API_PATH.EXTERNAL_EVENTS_LIST],
    queryFn: ({ signal }) => apiClient.getExternalEvents(signal),
  })

  return (
    <ContentWrapper size="xl" className="gap-6">
      <HydrationBoundary state={dehydrate(queryClient)}>
        <ExternalEventsPage />
      </HydrationBoundary>
    </ContentWrapper>
  )
}
