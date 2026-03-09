'use client'

import { useState, useMemo, useCallback } from 'react'
import Link from 'next/link'
import {
  useQuery,
  useMutation,
  useQueryClient,
  keepPreviousData,
} from '@tanstack/react-query'
import { CalendarPlus, Filter, Loader2, RotateCcw } from 'lucide-react'
import { format, subMonths, startOfMonth } from 'date-fns'
import {
  createParser,
  parseAsString,
  parseAsStringLiteral,
  useQueryStates,
} from 'nuqs'
import {
  API_PATH,
  apiClient,
  EVENT_TYPE_VALUES,
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
import toast from 'react-hot-toast'
import { useDebouncedState } from '@/hooks/use-debounced-state'
import { ActivistRegistry } from './activist-registry'

export type EventListMode = 'events' | 'connections'

const parseAsDateString = createParser({
  parse: (v) => (/^\d{4}-\d{2}-\d{2}$/.test(v) ? v : null),
  serialize: (v) => v,
})

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

// Activist autocomplete input: only commits a filter when a suggestion is
// explicitly selected. Typing alone has no effect on the query.
function ActivistFilterInput({
  registry,
  value,
  onChange,
  onSelect,
  onBlur,
  label,
  className,
}: {
  registry: ActivistRegistry
  value: string
  onChange: (v: string) => void
  onSelect: (v: string) => void
  onBlur: () => void
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
        onCommit={(meta) => {
          if (meta.fromSuggestion) onSelect(meta.value)
        }}
        onBlur={onBlur}
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

  // URL params drive the query and are the source of truth for committed filters
  const [urlParams, setUrlParams] = useQueryStates({
    name: parseAsString.withDefault(''),
    activist: parseAsString.withDefault(''),
    start: parseAsDateString.withDefault(defaultParams.event_date_start),
    end: parseAsDateString.withDefault(defaultParams.event_date_end),
    type: parseAsStringLiteral(EVENT_TYPE_VALUES).withDefault(
      defaultParams.event_type,
    ),
  })

  // Name input debounces into URL params
  const [nameInput, setNameInput] = useDebouncedState(urlParams.name, (v) =>
    setUrlParams({ name: v || null }),
  )

  // Activist input only commits to URL when a suggestion is explicitly selected.
  // On blur without selection, the input reverts to the committed URL value.
  const [activistInput, setActivistInput] = useState(urlParams.activist)
  const [prevActivistUrl, setPrevActivistUrl] = useState(urlParams.activist)
  if (urlParams.activist !== prevActivistUrl) {
    setPrevActivistUrl(urlParams.activist)
    setActivistInput(urlParams.activist)
  }

  const handleActivistBlur = () => {
    if (activistInput === '') {
      // User cleared the field — remove the filter
      setUrlParams({ activist: null })
    } else {
      // User typed but didn't select — revert display to the committed value
      setActivistInput(urlParams.activist)
    }
  }

  const [showFilters, setShowFilters] = useState(false)

  // Dirty = committed filters differ from defaults (controls Reset button)
  const isDirty =
    urlParams.name !== '' ||
    urlParams.activist !== '' ||
    urlParams.start !== defaultParams.event_date_start ||
    urlParams.end !== defaultParams.event_date_end ||
    urlParams.type !== defaultParams.event_type

  const handleReset = () => {
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
    event_type: isConnections ? 'Connection' : urlParams.type,
  }

  const {
    data: events,
    isLoading,
    isPlaceholderData,
    isError,
    error,
  } = useQuery({
    queryKey: [API_PATH.EVENT_LIST, committedParams],
    queryFn: () => apiClient.getEventList(committedParams),
    placeholderData: keepPreviousData,
  })

  const deleteMutation = useMutation({
    mutationFn: (eventId: number) => apiClient.deleteEvent(eventId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [API_PATH.EVENT_LIST] })
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to delete event. Please try again.')
    },
  })

  const handleDelete = useCallback(
    (event: EventListItem) => {
      const confirmed = window.confirm(
        `Are you sure you want to delete "${event.event_name}"?`,
      )
      if (confirmed) {
        deleteMutation.mutate(event.event_id)
      }
    },
    [deleteMutation],
  )

  const title = isConnections ? 'All Coachings' : 'All Events'
  const newHref = isConnections ? '/coachings/new' : '/events/new'
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
        <div className="flex items-center gap-3">
          <Button
            variant="outline"
            size="sm"
            className="md:hidden"
            onClick={() => setShowFilters((v) => !v)}
          >
            <Filter className="h-4 w-4" />
            {showFilters ? 'Hide filters' : 'Show filters'}
          </Button>
        </div>

        <div
          className={cn(
            'flex-col gap-4 rounded-md border p-4',
            showFilters ? 'flex' : 'hidden md:flex',
          )}
        >
          <div className="flex flex-col sm:flex-row sm:items-end gap-4">
            <div className="flex flex-col gap-1.5 sm:flex-1 sm:max-w-xs">
              <Label htmlFor="filter-name">
                {isConnections ? 'Coach' : 'Event Name'}
              </Label>
              <Input
                id="filter-name"
                value={nameInput}
                onChange={(e) => setNameInput(e.target.value)}
              />
            </div>

            <ActivistFilterInput
              registry={registry}
              value={activistInput}
              onChange={setActivistInput}
              onSelect={(v) => setUrlParams({ activist: v || null })}
              onBlur={handleActivistBlur}
              label={isConnections ? 'Coachee' : 'Activist'}
              className="sm:flex-1 sm:max-w-xs"
            />

            {!isConnections && (
              <div className="flex flex-col gap-1.5 sm:w-44">
                <Label>Type</Label>
                <Select
                  value={urlParams.type}
                  onValueChange={(v) => setUrlParams({ type: v as EventType })}
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
                value={
                  urlParams.start
                    ? new Date(urlParams.start + 'T00:00:00')
                    : undefined
                }
                onValueChange={(date) =>
                  setUrlParams({
                    start: date ? format(date, 'yyyy-MM-dd') : null,
                  })
                }
                placeholder="Start date"
                className="w-40"
              />
            </div>

            <div className="flex flex-col gap-1.5">
              <Label>To</Label>
              <DatePicker
                value={
                  urlParams.end
                    ? new Date(urlParams.end + 'T00:00:00')
                    : undefined
                }
                onValueChange={(date) =>
                  setUrlParams({
                    end: date ? format(date, 'yyyy-MM-dd') : null,
                  })
                }
                placeholder="End date"
                className="w-40"
              />
            </div>
          </div>

          {isDirty && (
            <div>
              <Button variant="ghost" size="sm" onClick={handleReset}>
                <RotateCcw className="h-4 w-4" />
                Reset filters
              </Button>
            </div>
          )}
        </div>
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
        <div
          className={cn(
            'flex flex-col gap-4 transition-opacity',
            isPlaceholderData && 'opacity-50',
          )}
        >
          <div className="text-sm text-muted-foreground">
            {events.length} {isConnections ? 'coaching' : 'event'}
            {events.length !== 1 ? 's' : ''} found
          </div>
          <EventListTable events={events} mode={mode} onDelete={handleDelete} />
        </div>
      )}
    </div>
  )
}
