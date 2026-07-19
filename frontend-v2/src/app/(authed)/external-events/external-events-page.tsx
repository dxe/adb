'use client'

import { useQuery } from '@tanstack/react-query'
import { Loader2 } from 'lucide-react'
import { API_PATH, apiClient } from '@/lib/api'
import { ExternalEventsTable } from './external-events-table'

export default function ExternalEventsPage() {
  const { data: events, isLoading } = useQuery({
    queryKey: [API_PATH.EXTERNAL_EVENTS_LIST],
    queryFn: ({ signal }) => apiClient.getExternalEvents(signal),
  })

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col gap-1">
        <h1 className="text-2xl font-semibold">Facebook Events</h1>
        <p className="text-muted-foreground text-sm">
          Feature or cancel upcoming Facebook/Eventbrite events displayed on the
          public events page.
        </p>
      </div>

      {isLoading ? (
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <Loader2 className="h-4 w-4 animate-spin" />
          Loading events...
        </div>
      ) : events ? (
        <ExternalEventsTable events={events} />
      ) : (
        <div className="text-sm text-destructive">Failed to load events</div>
      )}
    </div>
  )
}
