'use client'

import { useEffect, useMemo, useState } from 'react'
import Link from 'next/link'
import { useQuery } from '@tanstack/react-query'
import {
  ArrowRight,
  CalendarPlus,
  ChevronRight,
  Clock,
  Loader2,
  Users,
} from 'lucide-react'
import { API_PATH, apiClient, EventListParams } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { useAuthedPageContext } from '@/hooks/useAuthedPageContext'
import {
  formatEventTimeRange,
  getBrowserTimezone,
  isEventHappeningNow,
  todayInTimezone,
} from '@/lib/timezone'

export function HomeHub() {
  const { user } = useAuthedPageContext()

  // Re-render each minute so the "Happening now" badge appears/clears as time
  // passes, and so the page rolls over to the next day at local midnight,
  // without needing a manual refresh.
  const [now, setNow] = useState(() => new Date())
  useEffect(() => {
    const id = setInterval(() => setNow(new Date()), 60_000)
    return () => clearInterval(id)
  }, [])

  // Date/time output depends on the viewer's clock and timezone, which differ
  // between the server (SSR) and the browser — near midnight or across zones
  // they can land on different calendar days, causing a hydration mismatch.
  // Gate the date-dependent label on a mount flag so the server and first
  // client render agree; the real label fills in after hydration.
  const [mounted, setMounted] = useState(false)
  useEffect(() => setMounted(true), [])

  // "Today" is computed in the viewer's timezone (a defined zone) rather than
  // the DB/browser raw date, to avoid off-by-one-day errors. Chapters have no
  // timezone of their own, so the creator's zone is the best available default.
  // Tied to `now` so a tab left open past midnight rolls over automatically.
  const today = useMemo(() => todayInTimezone(getBrowserTimezone(), now), [now])

  // A friendly, human-readable form of today's date for the header. `today` is
  // already YYYY-MM-DD; parse at local midnight so the label can't drift a day.
  const todayLabel = useMemo(
    () =>
      new Date(`${today}T00:00:00`).toLocaleDateString(undefined, {
        weekday: 'long',
        month: 'long',
        day: 'numeric',
      }),
    [today],
  )

  // Upcoming = today through a year out. This reuses the existing events list
  // page via its URL date filter rather than duplicating a list here. The
  // events list keeps its own default (last month → today) when opened directly.
  const upcomingHref = useMemo(() => {
    // Add a year via Date math (not string slicing) so leap days like
    // 2028-02-29 don't produce an invalid 2029-02-29.
    const endDate = new Date(`${today}T00:00:00`)
    endDate.setFullYear(endDate.getFullYear() + 1)
    const end = [
      endDate.getFullYear(),
      String(endDate.getMonth() + 1).padStart(2, '0'),
      String(endDate.getDate()).padStart(2, '0'),
    ].join('-')
    return `/events?start=${today}&end=${end}`
  }, [today])

  const params: EventListParams = {
    event_date_start: today,
    event_date_end: today,
    event_type: 'noConnections',
  }

  const {
    data: events,
    isLoading,
    isError,
  } = useQuery({
    queryKey: [API_PATH.EVENT_LIST, params],
    queryFn: ({ signal }) => apiClient.getEventList(params, signal),
  })

  // Chronological by start time, with untimed events (quick attendance) last.
  const sortedEvents = useMemo(() => {
    if (!events) return events
    return [...events].sort((a, b) => {
      const aStart = a.start_time ?? ''
      const bStart = b.start_time ?? ''
      if (aStart && bStart) return aStart.localeCompare(bStart)
      if (aStart) return -1
      if (bStart) return 1
      return 0
    })
  }, [events])

  return (
    <div className="flex flex-col gap-6">
      {/* Greeting orients the user and sets the page apart from a bare list. */}
      <header>
        <h1 className="text-2xl font-semibold tracking-tight">
          Welcome{user.Name ? `, ${user.Name.split(' ')[0]}` : ''}
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">
          {user.ChapterName}
          {mounted && ` · ${todayLabel}`}
        </p>
      </header>

      {/* Today's events for quick attendance entry, plus a single entry point
          for creating any new event (public or not). */}
      <section className="rounded-xl border bg-card shadow-sm">
        <div className="flex items-start justify-between gap-3 border-b p-5">
          <div>
            <h2 className="text-lg font-semibold">Today&apos;s events</h2>
            <p className="mt-1 text-sm text-muted-foreground">
              Pick an event happening today to enter attendance, or create a new
              event.
            </p>
          </div>
          <Button asChild>
            <Link href="/events/new">
              <CalendarPlus className="h-4 w-4" />
              New event
            </Link>
          </Button>
        </div>

        <div className="flex flex-col gap-2 p-5">
          {isLoading && (
            <div className="flex items-center gap-2 py-6 text-sm text-muted-foreground">
              <Loader2 className="h-4 w-4 animate-spin" />
              Loading today&apos;s events...
            </div>
          )}

          {isError && (
            <div className="rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">
              Failed to load today&apos;s events.
            </div>
          )}

          {!isLoading && !isError && events && events.length === 0 && (
            <div className="rounded-md border border-dashed px-4 py-8 text-center text-sm text-muted-foreground">
              No events scheduled for today.
            </div>
          )}

          {!isLoading &&
            !isError &&
            sortedEvents &&
            sortedEvents.map((event) => {
              const timeRange = event.start_time
                ? formatEventTimeRange(
                    event.event_date,
                    event.start_time,
                    event.end_time ?? '',
                    event.timezone ?? '',
                  )
                : ''
              const happeningNow = isEventHappeningNow(
                event.event_date,
                event.start_time,
                event.end_time,
                event.timezone,
                now,
              )
              return (
                <Link
                  key={event.event_id}
                  href={`/events/${event.event_id}`}
                  className="group flex items-center justify-between gap-3 rounded-lg border px-4 py-3 transition-colors hover:border-primary/40 hover:bg-accent"
                >
                  <div className="flex min-w-0 flex-col">
                    <span className="flex items-center gap-2">
                      <span className="truncate font-medium">
                        {event.event_name}
                      </span>
                      {happeningNow && (
                        <span className="inline-flex shrink-0 items-center gap-1 rounded-full bg-green-100 px-2 py-0.5 text-xs font-medium text-green-700">
                          <span className="h-1.5 w-1.5 animate-pulse rounded-full bg-green-500" />
                          Happening now
                        </span>
                      )}
                    </span>
                    <span className="mt-0.5 flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
                      <span className="inline-flex items-center rounded-full bg-muted px-2 py-0.5 font-medium">
                        {event.event_type}
                      </span>
                      {timeRange && (
                        <span className="inline-flex items-center gap-1">
                          <Clock className="h-3 w-3" />
                          {timeRange}
                        </span>
                      )}
                      <span className="inline-flex items-center gap-1">
                        <Users className="h-3 w-3" />
                        {event.attendees.length}
                      </span>
                    </span>
                  </div>
                  <ChevronRight className="h-4 w-4 shrink-0 text-muted-foreground transition-transform group-hover:translate-x-0.5 group-hover:text-foreground" />
                </Link>
              )
            })}
        </div>

        <div className="border-t p-5">
          <Link
            href={upcomingHref}
            className="inline-flex items-center gap-1 text-sm font-medium text-primary hover:underline"
          >
            View today&apos;s &amp; upcoming events
            <ArrowRight className="h-4 w-4" />
          </Link>
        </div>
      </section>
    </div>
  )
}
