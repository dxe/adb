import { ContentWrapper } from '@/app/content-wrapper'
import { AuthedPageLayout } from '@/app/authed-page-layout'
import { EventForm } from '../../event/event-form'
import { Navbar } from '@/components/nav'
import { ApiClient, API_PATH } from '@/lib/api'
import {
  QueryClient,
  HydrationBoundary,
  dehydrate,
} from '@tanstack/react-query'
import { getCookies } from '@/lib/auth'

type EditCoachingPageProps = {
  params: Promise<{ id: string }>
}

export default async function EditCoachingPage({
  params,
}: EditCoachingPageProps) {
  const { id } = await params
  const apiClient = new ApiClient(await getCookies())
  const queryClient = new QueryClient()

  // Prefetch event data during SSR
  await queryClient.prefetchQuery({
    queryKey: [API_PATH.EVENT_GET, id],
    queryFn: () => apiClient.getEvent(Number(id)),
  })

  return (
    <AuthedPageLayout pageName="EditConnection_beta">
      <HydrationBoundary state={dehydrate(queryClient)}>
        <Navbar />
        <ContentWrapper size="md" className="gap-8">
          <h1 className="text-3xl font-bold">Coaching</h1>
          <EventForm mode="connection" />
        </ContentWrapper>
      </HydrationBoundary>
    </AuthedPageLayout>
  )
}
