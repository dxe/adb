import {
  dehydrate,
  HydrationBoundary,
  QueryClient,
} from '@tanstack/react-query'
import { notFound } from 'next/navigation'
import { ContentWrapper } from '@/app/content-wrapper'
import { EventConfirmation } from './event-confirmation'
import { API_PATH, ApiClient } from '@/lib/api'
import { getCookies } from '@/lib/auth'
import { redirectForHttpError } from '@/lib/server-auth'

// Shown right after a scheduled (public) event is created. Attendance happens
// later, at the event, so we confirm the event is on the schedule and offer the
// natural next steps instead of dropping the user on the attendance page.
export default async function EventConfirmationPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = await params
  const eventId = parseInt(id)
  if (Number.isNaN(eventId)) {
    notFound()
  }

  const apiClient = new ApiClient(await getCookies())
  const queryClient = new QueryClient()

  await redirectForHttpError(() =>
    // Intentionally use fetchQuery instead of prefetchQuery; see redirectForHttpError for details.
    queryClient.fetchQuery({
      queryKey: [API_PATH.EVENT_GET, String(eventId)],
      queryFn: ({ signal }) => apiClient.getEvent(eventId, signal),
    }),
  )

  return (
    <ContentWrapper size="sm" className="gap-8">
      <HydrationBoundary state={dehydrate(queryClient)}>
        <EventConfirmation eventId={eventId} />
      </HydrationBoundary>
    </ContentWrapper>
  )
}
