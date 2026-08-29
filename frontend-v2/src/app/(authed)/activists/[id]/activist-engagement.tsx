'use client'

import { useQuery } from '@tanstack/react-query'
import { CalendarDays, MessageSquare } from 'lucide-react'
import {
  API_PATH,
  apiClient,
  type ActivistTimelineItem,
  type ActivistTimelineEventPayload,
  type ActivistTimelineInteractionPayload,
} from '@/lib/api'
import { Badge } from '@/components/ui/badge'
import {
  formatInstantTimeForActivists,
  formatYmdForActivists,
} from '../date-time'

export function useActivistTimeline(activistId: number) {
  return useQuery({
    queryKey: [API_PATH.ACTIVIST_TIMELINE, activistId],
    queryFn: ({ signal }) => apiClient.getActivistTimeline(activistId, signal),
  })
}

// A single chronological history of everything the chapter has recorded about
// an activist: events they attended and interactions organizers logged with
// them, newest first. The server merges and caps the two sources, so this
// renders the items in the order it receives them.
export function ActivistEngagement({ activistId }: { activistId: number }) {
  const { data, isError, isLoading } = useActivistTimeline(activistId)

  if (isLoading) {
    return (
      <div className="animate-pulse text-sm text-muted-foreground">
        Loading engagement history...
      </div>
    )
  }
  if (isError || !data) {
    return (
      <div className="text-sm text-muted-foreground">
        Unable to load engagement history.
      </div>
    )
  }
  if (data.items.length === 0) {
    return (
      <p className="text-sm text-muted-foreground italic">
        No events attended or interactions logged yet.
      </p>
    )
  }

  return (
    <div className="flex flex-col gap-4">
      <ol className="flex flex-col">
        {data.items.map((item) => (
          <TimelineRow key={`${item.type}-${item.id}`} item={item} />
        ))}
      </ol>
      {data.truncated && (
        <p className="text-xs text-muted-foreground italic">
          Showing the most recent {data.items.length} items. Older history is
          not shown.
        </p>
      )}
    </div>
  )
}

function TimelineRow({ item }: { item: ActivistTimelineItem }) {
  const isEvent = item.type === 'event'
  const Icon = isEvent ? CalendarDays : MessageSquare
  // Every item carries an instant, but only some carry a real time of day: an
  // event with no recorded start time gets a placeholder one so it sorts
  // within its day, and showing that would invent a time the event never had.
  const time = item.has_time
    ? formatInstantTimeForActivists(item.timestamp)
    : ''

  return (
    <li className="group flex gap-3">
      {/* Marker column: an icon per item, joined by a line that runs the rest
          of the row's height so consecutive rows form one continuous rail. */}
      <div className="flex flex-col items-center">
        <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full border bg-card text-muted-foreground">
          <Icon className="h-3.5 w-3.5" aria-hidden="true" />
        </span>
        <span className="w-px flex-1 bg-border group-last:hidden" />
      </div>

      <div className="min-w-0 flex-1 pb-6 group-last:pb-0">
        <p className="text-xs text-muted-foreground">
          {formatYmdForActivists(item.date)}
          {time && ` · ${time}`}
        </p>
        {isEvent ? (
          <EventBody event={item.event} />
        ) : (
          <InteractionBody interaction={item.interaction} />
        )}
      </div>
    </li>
  )
}

function EventBody({ event }: { event: ActivistTimelineEventPayload }) {
  return (
    <div className="flex flex-wrap items-center gap-x-2 gap-y-1">
      <span className="font-medium">{event.name || 'Event'}</span>
      {event.type && (
        <Badge variant="secondary" className="font-normal">
          {event.type}
        </Badge>
      )}
    </div>
  )
}

function InteractionBody({
  interaction,
}: {
  interaction: ActivistTimelineInteractionPayload
}) {
  return (
    <div className="flex flex-col gap-1">
      <div className="flex flex-wrap items-center gap-x-2 gap-y-1">
        <span className="font-medium">
          {interaction.method || 'Interaction'}
        </span>
        {interaction.user_name && (
          <span className="text-sm text-muted-foreground">
            by {interaction.user_name}
          </span>
        )}
        {interaction.outcome && (
          <Badge variant="outline" className="font-normal">
            {interaction.outcome}
          </Badge>
        )}
      </div>
      {interaction.notes && (
        <p className="whitespace-pre-wrap text-sm text-muted-foreground">
          {interaction.notes}
        </p>
      )}
    </div>
  )
}
