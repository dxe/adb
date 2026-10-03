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

  // prefetchQuery, not fetchQuery + redirectForHttpError: the endpoint isn't
  // admin-gated, so a failed prefetch just falls through to the client useQuery.
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
