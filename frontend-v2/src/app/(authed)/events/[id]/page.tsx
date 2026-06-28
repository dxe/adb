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
  searchParams,
}: {
  params: Promise<{ id: string }>
  searchParams: Promise<{ expanded?: string; attendees?: string }>
}) {
  const [{ id }, { expanded, attendees }] = await Promise.all([
    params,
    searchParams,
  ])
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
        <h1 className="text-3xl font-bold">Event</h1>
        {/* Opening an event to manage it (event list, home, "Take attendance
            now") shows attendees outright. Only the confirmation page's "Edit
            event" link, which targets a freshly created advance event, passes
            attendees=0 to keep them tucked behind the "Add attendees" link. */}
        <EventForm
          mode="event"
          startExpanded={expanded === '1'}
          startAttendeesExpanded={attendees !== '0'}
        />
      </HydrationBoundary>
    </ContentWrapper>
  )
}
