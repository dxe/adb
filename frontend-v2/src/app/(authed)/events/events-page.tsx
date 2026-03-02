'use client'

import { useState, KeyboardEvent } from 'react'
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
import { Popover, PopoverAnchor, PopoverContent } from '@/components/ui/popover'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { DatePicker } from '@/components/ui/date-picker'
import { useActivistRegistry } from './useActivistRegistry'
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
  value,
  onChange,
  onEnter,
  label,
}: {
  value: string
  onChange: (v: string) => void
  onEnter: () => void
  label: string
}) {
  const { registry } = useActivistRegistry()
  const [suggestions, setSuggestions] = useState<string[]>([])
  const [selectedIndex, setSelectedIndex] = useState(-1)
  const [focused, setFocused] = useState(false)
  const open = focused && suggestions.length > 0

  const handleChange = (v: string) => {
    onChange(v)
    setSuggestions(registry.getSuggestions(v))
    setSelectedIndex(-1)
  }

  const handleSelect = (name: string) => {
    onChange(name)
    setSuggestions([])
    setSelectedIndex(-1)
  }

  const handleKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      setSelectedIndex((i) => (i === suggestions.length - 1 ? 0 : i + 1))
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      setSelectedIndex((i) => (i <= 0 ? suggestions.length - 1 : i - 1))
    } else if (e.key === 'Escape') {
      setSuggestions([])
      setSelectedIndex(-1)
    } else if (e.key === 'Enter') {
      e.preventDefault()
      if (selectedIndex >= 0 && selectedIndex < suggestions.length) {
        handleSelect(suggestions[selectedIndex])
      } else {
        onEnter()
      }
    }
  }

  return (
    <div className="flex flex-col gap-1.5">
      <Label htmlFor="filter-activist">{label}</Label>
      <Popover open={open}>
        <PopoverAnchor asChild>
          <Input
            id="filter-activist"
            value={value}
            onChange={(e) => handleChange(e.target.value)}
            onKeyDown={handleKeyDown}
            onFocus={() => setFocused(true)}
            onBlur={() => {
              setFocused(false)
              setSuggestions([])
              setSelectedIndex(-1)
            }}
            className="w-48"
            autoComplete="off"
          />
        </PopoverAnchor>
        <PopoverContent
          className="p-0 w-[var(--radix-popover-trigger-width)]"
          align="start"
          sideOffset={4}
          onOpenAutoFocus={(e) => e.preventDefault()}
          onCloseAutoFocus={(e) => e.preventDefault()}
        >
          <ul className="max-h-[300px] overflow-y-auto rounded-md border border-gray-200 bg-white shadow-lg">
            {suggestions.map((name, i) => (
              <li
                key={name}
                className={cn(
                  'cursor-pointer px-3 py-1 hover:bg-gray-100 text-sm',
                  i === selectedIndex ? 'bg-neutral-100' : '',
                )}
                onMouseDown={(e) => {
                  e.preventDefault()
                  handleSelect(name)
                }}
              >
                {name}
              </li>
            ))}
          </ul>
        </PopoverContent>
      </Popover>
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
  const defaultParams = buildDefaultParams(mode)
  const isConnections = mode === 'connections'
  const queryClient = useQueryClient()

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
      start: formStart !== defaultParams.event_date_start ? formStart : null,
      end: formEnd !== defaultParams.event_date_end ? formEnd : null,
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
      `Are you sure you want to delete "${event.event_name}"?\n\nPress OK to delete this event.`,
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
          <div className="flex flex-wrap items-end gap-4 rounded-md border p-4">
            {/* Event Name — full width on mobile so activist+type share the next row */}
            <div className="flex flex-col gap-1.5 w-full sm:w-auto">
              <Label htmlFor="filter-name">
                {isConnections ? 'Coach' : 'Event Name'}
              </Label>
              <Input
                id="filter-name"
                value={formName}
                onChange={(e) => setFormName(e.target.value)}
                className="w-full sm:w-48"
                onKeyDown={(e) => e.key === 'Enter' && handleFilter()}
              />
            </div>

            {/* Activist + Type grouped so they wrap together */}
            <div className="flex items-end gap-4">
              <ActivistFilterInput
                value={formActivist}
                onChange={setFormActivist}
                onEnter={handleFilter}
                label={isConnections ? 'Coachee' : 'Activist'}
              />

              {!isConnections && (
                <div
                  className="flex flex-col gap-1.5 shrink-0"
                  style={{ width: '11rem', minWidth: 0 }}
                >
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

            <div className="flex items-end gap-4">
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
