'use client'

import { useState, useMemo } from 'react'
import Link from 'next/link'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { CalendarPlus, Filter, Loader2, RotateCcw } from 'lucide-react'
import { parseISO, format, subMonths, startOfMonth } from 'date-fns'
import { parseAsString, useQueryStates } from 'nuqs'
import {
  API_PATH,
  apiClient,
  EventListItem,
  EventListParams,
  EventType,
} from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { DatePicker } from '@/components/ui/date-picker'
import { useActivistRegistry } from './useActivistRegistry'
import { SuggestionInput } from './suggestion-input'
import { EventListTable } from './event-list-table'
import { cn } from '@/lib/utils'

export type EventListMode = 'events' | 'connections'

const EVENT_TYPES: { value: EventType; label: string }[] = [
  { value: 'noConnections', label: 'All' },
  { value: 'Action', label: 'Action' },
  { value: 'Campaign Action', label: 'Campaign Action' },
  { value: 'Community', label: 'Community' },
  { value: 'Frontline Surveillance', label: 'Frontline Surveillance' },
  { value: 'Meeting', label: 'Meeting' },
  { value: 'Outreach', label: 'Outreach' },
  { value: 'Animal Care', label: 'Animal Care' },
  { value: 'Training', label: 'Training' },
  { value: 'mpiDA', label: 'MPI: Direct Action' },
  { value: 'mpiCOM', label: 'MPI: Community' },
]

// Activist autocomplete input backed by the activist registry
function ActivistFilterInput({
  registry,
  value,
  onChange,
  onEnter,
  label,
  className,
}: {
  registry: ReturnType<typeof useActivistRegistry>['registry']
  value: string
  onChange: (v: string) => void
  onEnter: () => void
  label: string
  className?: string
}) {
  return (
    <div className={cn('flex flex-col gap-1.5', className)}>
      <Label htmlFor="filter-activist">{label}</Label>
      <SuggestionInput
        id="filter-activist"
        value={value}
        onValueChange={onChange}
        getSuggestions={(v) => registry.getSuggestions(v)}
        onCommit={({ key }) => {
          if (key === 'Enter') {
            onEnter()
          }
        }}
        className="w-full"
        size="sm"
      />
    </div>
  )
}

function buildDefaultParams(mode: EventListMode): EventListParams {
  const today = new Date()
  return {
    event_date_start: format(startOfMonth(subMonths(today, 1)), 'yyyy-MM-dd'),
    event_date_end: format(today, 'yyyy-MM-dd'),
    event_type: mode === 'connections' ? 'Connection' : 'noConnections',
  }
}

type Props = {
  mode?: EventListMode
}

export default function EventsPage({ mode = 'events' }: Props) {
  const defaultParams = useMemo(() => buildDefaultParams(mode), [mode])
  const isConnections = mode === 'connections'
  const queryClient = useQueryClient()
  const { registry } = useActivistRegistry()

  // URL params = committed filters (drives the query + shareable link)
  const [urlParams, setUrlParams] = useQueryStates({
    name: parseAsString.withDefault(''),
    activist: parseAsString.withDefault(''),
    start: parseAsString.withDefault(defaultParams.event_date_start),
    end: parseAsString.withDefault(defaultParams.event_date_end),
    type: parseAsString.withDefault(defaultParams.event_type),
  })

  // Form state — local, uncommitted until Filter is clicked
  // Initialized from URL so shared links pre-populate the form
  const [formName, setFormName] = useState(urlParams.name)
  const [formActivist, setFormActivist] = useState(urlParams.activist)
  const [formStart, setFormStart] = useState(urlParams.start)
  const [formEnd, setFormEnd] = useState(urlParams.end)
  const [formType, setFormType] = useState<EventType>(
    urlParams.type as EventType,
  )
  const [showFilters, setShowFilters] = useState(false)

  // Dirty = committed filters differ from defaults (controls Reset button)
  const isDirty =
    urlParams.name !== '' ||
    urlParams.activist !== '' ||
    urlParams.start !== defaultParams.event_date_start ||
    urlParams.end !== defaultParams.event_date_end ||
    urlParams.type !== defaultParams.event_type

  const handleFilter = () => {
    setUrlParams({
      name: formName || null,
      activist: formActivist || null,
      start:
        formStart && formStart !== defaultParams.event_date_start
          ? formStart
          : null,
      end: formEnd && formEnd !== defaultParams.event_date_end ? formEnd : null,
      type: formType !== defaultParams.event_type ? formType : null,
    })
  }

  const handleReset = () => {
    setFormName('')
    setFormActivist('')
    setFormStart(defaultParams.event_date_start)
    setFormEnd(defaultParams.event_date_end)
    setFormType(defaultParams.event_type as EventType)
    setUrlParams({
      name: null,
      activist: null,
      start: null,
      end: null,
      type: null,
    })
  }

  const committedParams: EventListParams = {
    event_name: urlParams.name || undefined,
    event_activist: urlParams.activist || undefined,
    event_date_start: urlParams.start,
    event_date_end: urlParams.end,
    event_type: isConnections ? 'Connection' : (urlParams.type as EventType),
  }

  const {
    data: events,
    isLoading,
    isError,
    error,
  } = useQuery({
    queryKey: [API_PATH.EVENT_LIST, committedParams],
    queryFn: () => apiClient.getEventList(committedParams),
  })

  const deleteMutation = useMutation({
    mutationFn: (eventId: number) => apiClient.deleteEvent(eventId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [API_PATH.EVENT_LIST] })
    },
  })

  const handleDelete = (event: EventListItem) => {
    const confirmed = window.confirm(
      `Are you sure you want to delete "${event.event_name}"?`,
    )
    if (confirmed) {
      deleteMutation.mutate(event.event_id)
    }
  }

  const title = isConnections ? 'All Coachings' : 'All Events'
  const newHref = isConnections ? '/coaching/new' : '/events/new'
  const newLabel = isConnections ? 'New Coaching' : 'New Event'

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-start justify-between gap-3">
        <h1 className="text-2xl font-semibold">{title}</h1>
        <Button asChild>
          <Link href={newHref}>
            <CalendarPlus className="h-4 w-4" />
            {newLabel}
          </Link>
        </Button>
      </div>

      <div className="flex flex-col gap-4">
        <div>
          <Button
            variant="outline"
            size="sm"
            onClick={() => setShowFilters((v) => !v)}
          >
            <Filter className="h-4 w-4" />
            {showFilters ? 'Hide filters' : 'Show filters'}
          </Button>
        </div>

        {showFilters && (
          <div className="flex flex-col gap-4 rounded-md border p-4">
            <div className="flex flex-col sm:flex-row sm:items-end gap-4">
              <div className="flex flex-col gap-1.5 sm:flex-1 sm:max-w-xs">
                <Label htmlFor="filter-name">
                  {isConnections ? 'Coach' : 'Event Name'}
                </Label>
                <Input
                  id="filter-name"
                  value={formName}
                  onChange={(e) => setFormName(e.target.value)}
                  onKeyDown={(e) => e.key === 'Enter' && handleFilter()}
                />
              </div>

              <ActivistFilterInput
                registry={registry}
                value={formActivist}
                onChange={setFormActivist}
                onEnter={handleFilter}
                label={isConnections ? 'Coachee' : 'Activist'}
                className="sm:flex-1 sm:max-w-xs"
              />

              {!isConnections && (
                <div className="flex flex-col gap-1.5 sm:w-44">
                  <Label>Type</Label>
                  <Select
                    value={formType}
                    onValueChange={(v) => setFormType(v as EventType)}
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {EVENT_TYPES.map(({ value, label }) => (
                        <SelectItem key={value} value={value}>
                          {label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              )}
            </div>

            <div className="flex flex-wrap items-end gap-4">
              <div className="flex flex-col gap-1.5">
                <Label>From</Label>
                <DatePicker
                  value={formStart ? parseISO(formStart) : undefined}
                  onValueChange={(date) =>
                    setFormStart(date ? format(date, 'yyyy-MM-dd') : '')
                  }
                  placeholder="Start date"
                  className="w-40"
                />
              </div>

              <div className="flex flex-col gap-1.5">
                <Label>To</Label>
                <DatePicker
                  value={formEnd ? parseISO(formEnd) : undefined}
                  onValueChange={(date) =>
                    setFormEnd(date ? format(date, 'yyyy-MM-dd') : '')
                  }
                  placeholder="End date"
                  className="w-40"
                />
              </div>

              <div className="flex gap-2">
                <Button onClick={handleFilter}>Filter</Button>
                {isDirty && (
                  <Button variant="ghost" onClick={handleReset}>
                    <RotateCcw className="h-4 w-4" />
                    Reset filters
                  </Button>
                )}
              </div>
            </div>
          </div>
        )}
      </div>

      {isLoading && (
        <div className="flex items-center gap-2 text-muted-foreground text-sm">
          <Loader2 className="h-4 w-4 animate-spin" />
          Loading {isConnections ? 'coachings' : 'events'}...
        </div>
      )}

      {isError && (
        <div className="text-sm text-destructive">
          {error instanceof Error
            ? error.message
            : `Failed to load ${isConnections ? 'coachings' : 'events'}. Please try again.`}
        </div>
      )}

      {!isLoading && !isError && events && (
        <>
          <div className="text-sm text-muted-foreground">
            {events.length} {isConnections ? 'coaching' : 'event'}
            {events.length !== 1 ? 's' : ''} found
          </div>
          <EventListTable events={events} mode={mode} onDelete={handleDelete} />
        </>
      )}
    </div>
  )
}
