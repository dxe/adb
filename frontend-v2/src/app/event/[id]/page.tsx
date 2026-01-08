import { ContentWrapper } from '@/app/content-wrapper'
import { AuthedPageLayout } from '@/app/authed-page-layout'
import { EventForm } from '../event-form'
import { Navbar } from '@/components/nav'
import { ApiClient, API_PATH } from '@/lib/api'
import {
  QueryClient,
  HydrationBoundary,
  dehydrate,
} from '@tanstack/react-query'
import { getCookies } from '@/lib/auth'

type EditEventPageProps = {
  params: Promise<{ id: string }>
}

export default async function EditEventPage({ params }: EditEventPageProps) {
  const { id } = await params
  const apiClient = new ApiClient(await getCookies())
  const queryClient = new QueryClient()

  // Prefetch event data during SSR
  await queryClient.prefetchQuery({
    queryKey: [API_PATH.EVENT_GET, id],
    queryFn: () => apiClient.getEvent(Number(id)),
  })

  return (
    <AuthedPageLayout pageName="EditEvent_beta">
      <HydrationBoundary state={dehydrate(queryClient)}>
        <Navbar />
        <ContentWrapper size="md" className="gap-8">
          <h1 className="text-3xl font-bold">Attendance</h1>
          <EventForm mode="event" />
        </ContentWrapper>
      </HydrationBoundary>
    </AuthedPageLayout>
  )
}
