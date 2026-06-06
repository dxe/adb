'use client'

import Link from 'next/link'
import { useQuery } from '@tanstack/react-query'
import { format, parseISO } from 'date-fns'
import {
  CheckCircle2,
  Users,
  Pencil,
  CalendarPlus,
  Home,
  Clock,
  MapPin,
  Globe,
} from 'lucide-react'
import { API_PATH, apiClient } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { formatEventTimeRange } from '@/lib/time'

export function EventConfirmation({ eventId }: { eventId: number }) {
  // Hydrated by the server page (same query key as the event form), so this
  // resolves immediately on first render.
  const { data: event } = useQuery({
    queryKey: [API_PATH.EVENT_GET, String(eventId)],
    queryFn: ({ signal }) => apiClient.getEvent(eventId, signal),
  })

  const timeRange = event ? formatEventTimeRange(event) : ''
  const locationLabel =
    event?.location?.name || event?.location?.formatted_address

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col items-center gap-3 text-center">
        <CheckCircle2 className="h-12 w-12 text-green-600" />
        <div className="flex flex-col gap-1">
          <h2 className="text-xl font-semibold">Event created</h2>
          <p className="text-sm text-muted-foreground">
            Your event is scheduled.
          </p>
        </div>
      </div>

      {event && (
        <div className="flex flex-col gap-2 rounded-xl border bg-card p-5 shadow-sm">
          <p className="text-base font-semibold">{event.event_name}</p>
          <div className="flex flex-wrap items-center gap-x-3 gap-y-1.5 text-sm text-muted-foreground">
            <span className="inline-flex items-center rounded-full bg-muted px-2 py-0.5 text-xs font-medium">
              {event.event_type}
            </span>
            <span>{format(parseISO(event.event_date), 'PPP')}</span>
            {timeRange && (
              <span className="inline-flex items-center gap-1">
                <Clock className="h-3.5 w-3.5" />
                {timeRange}
              </span>
            )}
            {event.is_online ? (
              <span className="inline-flex items-center gap-1">
                <Globe className="h-3.5 w-3.5" />
                Online
              </span>
            ) : (
              locationLabel && (
                <span className="inline-flex items-center gap-1">
                  <MapPin className="h-3.5 w-3.5" />
                  {locationLabel}
                </span>
              )
            )}
          </div>
        </div>
      )}

      <div className="flex flex-col gap-2">
        <Button asChild>
          <Link href={`/events/${eventId}`}>
            <Users className="h-4 w-4" />
            Take attendance now
          </Link>
        </Button>
        <Button asChild variant="outline">
          <Link href={`/events/${eventId}?expanded=1`}>
            <Pencil className="h-4 w-4" />
            Edit event
          </Link>
        </Button>
        <Button asChild variant="outline">
          <Link href="/events/new">
            <CalendarPlus className="h-4 w-4" />
            Create another event
          </Link>
        </Button>
        <Button asChild variant="ghost">
          <Link href="/home">
            <Home className="h-4 w-4" />
            Done
          </Link>
        </Button>
      </div>
    </div>
  )
}
