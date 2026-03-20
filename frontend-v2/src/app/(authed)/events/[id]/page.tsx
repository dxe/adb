import {
  dehydrate,
  HydrationBoundary,
  QueryClient,
} from '@tanstack/react-query'
import { notFound } from 'next/navigation'
import { ContentWrapper } from '@/app/content-wrapper'
import { EventForm } from '../event-form'
import { API_PATH, ApiClient } from '@/lib/api'
import { getCookies } from '@/lib/auth'
import { redirectForHttpError } from '@/lib/server-auth'

export default async function EditEventPage({
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
        <h1 className="text-3xl font-bold">Attendance</h1>
        <EventForm mode="event" />
      </HydrationBoundary>
    </ContentWrapper>
  )
}
