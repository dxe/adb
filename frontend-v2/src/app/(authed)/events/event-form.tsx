'use client'

import { useRef, useState, useEffect, useMemo } from 'react'
import { useForm, useStore } from '@tanstack/react-form'
import { z } from 'zod'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { cn } from '@/lib/utils'
import { API_PATH, apiClient } from '@/lib/api'
import { useQuery, useQueryClient, useMutation } from '@tanstack/react-query'
import { useParams, useRouter } from 'next/navigation'
import toast from 'react-hot-toast'
import { useAuthedPageContext } from '@/hooks/useAuthedPageContext'
import { SF_BAY_CHAPTER_ID } from '@/lib/constants'
import { AttendeeInputField } from './attendee-input-field'
import { useActivistRegistry } from './useActivistRegistry'
import { PlacesAutocomplete } from './places-autocomplete'
import { DatePicker } from '@/components/ui/date-picker'
import { format, parseISO } from 'date-fns'
import { Save, ChevronDown, ChevronUp } from 'lucide-react'
import { TimeField } from '@/components/ui/time-field'
import {
  getBrowserTimezone,
  getCommonTimezones,
  getZoneAbbreviation,
  todayInTimezone,
} from '@/lib/timezone'

const EVENT_TYPES = [
  'Action',
  'Campaign Action',
  'Community',
  'Frontline Surveillance',
  'Meeting',
  'Outreach',
  'Animal Care',
  'Training',
] as const

const DEFAULT_FIELD_COUNT = 5
const MIN_EMPTY_FIELDS = 1

// Zod schema for form validation.
const attendeeSchema = z.object({
  name: z.string().refine(
    (name) => {
      const trimmed = name.trim()
      // Empty is ok, as it will be filtered out.
      if (trimmed === '') return true
      // Must have at least first and last name (contains a space).
      return trimmed.indexOf(' ') !== -1
    },
    {
      message: 'First & last name are required',
    },
  ),
})

const formSchema = z
  .object({
    eventName: z.string().min(1, 'Event name is required'),
    eventType: z.string().min(1, 'Event type is required'),
    eventDate: z.string().min(1, 'Event date is required'),
    suppressSurvey: z.boolean(),
    attendees: z.array(attendeeSchema),
    // Upcoming-event fields. Kept permissive here (empty string / false is
    // valid) so the plain attendance form still passes; the stricter per-mode
    // rules — a start time and a location for public events — are enforced by
    // the refinements below.
    isPublic: z.boolean(),
    isOnline: z.boolean(),
    description: z.string(),
    startTime: z.string(),
    endTime: z.string(),
    timezone: z.string(),
    googlePlaceId: z.string(),
    locationName: z.string(),
    formattedAddress: z.string(),
    lat: z.number().optional(),
    lng: z.number().optional(),
  })
  // TODO: events that cross midnight (end before start, e.g. an overnight
  // vigil) can't be expressed yet — leave the end time blank for now. If this
  // becomes a real need, add an explicit "end date" rather than inferring it.
  .refine((v) => !(v.startTime && v.endTime) || v.endTime >= v.startTime, {
    message: 'End time must be after start time',
    path: ['endTime'],
  })
  // Publicly listed events are scheduled in advance, so a start time is required.
  .refine((v) => !v.isPublic || Boolean(v.startTime), {
    message: 'Start time is required for public events',
    path: ['startTime'],
  })
  // In-person public events need a location; online ones don't.
  .refine((v) => !v.isPublic || v.isOnline || Boolean(v.googlePlaceId), {
    message: 'Location is required for in-person public events',
    path: ['formattedAddress'],
  })

type FormValues = z.infer<typeof formSchema>

export type EventFormMode = 'event' | 'connection'

type EventFormProps = {
  mode: EventFormMode
  // When editing a saved event, start with the detail fields expanded rather
  // than collapsed behind the summary bar. Used by the post-create confirmation
  // page's "Edit event" link (?expanded=1) so the user lands ready to fix
  // details.
  startExpanded?: boolean
}

export const EventForm = ({ mode, startExpanded }: EventFormProps) => {
  const router = useRouter()
  const params = useParams()
  const queryClient = useQueryClient()
  const { user, googlePlacesApiKey } = useAuthedPageContext()
  const eventId = params.id ? String(params.id) : undefined
  const isConnection = mode === 'connection'
  const browserTz = useMemo(() => getBrowserTimezone(), [])
  const timezones = useMemo(() => getCommonTimezones(), [])

  const inputRefs = useRef<(HTMLInputElement | null)[]>(
    Array(DEFAULT_FIELD_COUNT).fill(null),
  )
  const [activeInputIndex, setActiveInputIndex] = useState(0)

  // Initialize activist registry with IndexedDB caching and incremental sync.
  // The registry loads cached data from IndexedDB first, then syncs any
  // new/updated activists from the server in the background.
  const { registry: activistRegistry, isLoading: isLoadingActivists } =
    useActivistRegistry(user.ChapterID)

  // Fetch existing event/connection, if editing.
  const {
    data: eventData,
    isLoading: isEventLoading,
    isError: isEventError,
  } = useQuery({
    queryKey: [API_PATH.EVENT_GET, eventId],
    queryFn: ({ signal }) => apiClient.getEvent(Number(eventId), signal),
    enabled: !!eventId,
  })

  // The scheduled-event fields (time, location, description) are revealed by the
  // "Public event" checkbox — see showUpcomingFields below. We also surface them
  // when editing an event that already carries this data, so any pre-existing
  // event stays fully editable even if its Public box is unchecked.
  const editingHasUpcomingData = Boolean(
    eventData &&
    (eventData.is_public ||
      eventData.is_online ||
      eventData.start_time ||
      eventData.location?.google_place_id ||
      eventData.description),
  )

  // When editing a saved event the detail fields (name, type, date, schedule,
  // location, description) are usually already set, so they collapse behind a
  // summary header to keep the attendee list near the top for taking
  // attendance. New events start expanded since you're filling them out.
  const [detailsExpanded, setDetailsExpanded] = useState(
    startExpanded ?? !eventId,
  )

  const saveEventMutation = useMutation({
    mutationFn: isConnection ? apiClient.saveCoaching : apiClient.saveEvent,
    onSuccess: (result, variables) => {
      toast.success(`${isConnection ? 'Connection' : 'Event'} saved!`)

      // A brand-new scheduled (public) event is created in advance, so jumping
      // straight to attendance is awkward. Route to a confirmation page with
      // next-step choices instead, and skip the form-reset/routing below.
      if (!eventId && variables.is_public) {
        // Surface the new event in every list (home's "today", the events page)
        // without a manual refresh.
        queryClient.invalidateQueries({ queryKey: [API_PATH.EVENT_LIST] })
        router.push(`/events/${result.event_id}/confirmation`)
        return
      }

      if (!eventId) {
        const target = isConnection
          ? `/coachings/${result.event_id}`
          : `/events/${result.event_id}`
        router.push(target)
      }

      // Collapse the detail fields once saved, returning focus to the attendee
      // list. (New events collapse anyway via the remount above, but this also
      // covers editing a saved event with the details expanded.)
      setDetailsExpanded(false)

      // Reset the form's dirty state after successful save.
      // This prevents "unsaved changes" warning after successful save.

      // Use the attendee order from the current form state, not from the
      // server response (result.attendees). The server returns attendees in
      // arbitrary database order, but we want to preserve the user's input order.
      // This matches the behavior of the legacy Vue version.
      const savedAttendeeNames = form.state.values.attendees
        .map((a) => (a.name || '').trim())
        .filter((n) => n !== '')

      const current = form.state.values
      const newValues = {
        eventName: variables.event_name,
        eventType: variables.event_type,
        eventDate: variables.event_date,
        suppressSurvey: variables.suppress_survey,
        attendees: savedAttendeeNames
          .map((name) => ({ name }))
          .concat(
            Array(MIN_EMPTY_FIELDS)
              .fill(null)
              .map(() => ({ name: '' })),
          ),
        isPublic: current.isPublic,
        isOnline: current.isOnline,
        description: current.description,
        startTime: current.startTime,
        endTime: current.endTime,
        timezone: current.timezone,
        googlePlaceId: current.googlePlaceId,
        locationName: current.locationName,
        formattedAddress: current.formattedAddress,
        lat: current.lat,
        lng: current.lng,
      }

      // Use keepDefaultValues to work around TanStack Form bug:
      // https://github.com/TanStack/form/issues/1798
      form.reset(newValues, { keepDefaultValues: true })

      // Manually update defaultValues since keepDefaultValues prevents it.
      form.options.defaultValues = newValues

      // Refresh activist list to include newly created activists.
      // This ensures they appear in autocomplete suggestions.
      queryClient.invalidateQueries({
        queryKey: [API_PATH.ACTIVIST_LIST_BASIC],
      })
      queryClient.invalidateQueries({
        queryKey: [API_PATH.EVENT_GET, eventId],
      })
      // Refresh every event list so the change shows without a manual reload.
      queryClient.invalidateQueries({
        queryKey: [API_PATH.EVENT_LIST],
      })
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Error saving event')
    },
  })

  const initialValues: FormValues = useMemo(() => {
    return {
      eventName: eventData?.event_name || '',
      eventType: eventData?.event_type || (isConnection ? 'Connection' : ''),
      eventDate: eventData?.event_date || '',
      // For new events, non-SF Bay chapters default to not sending surveys.
      suppressSurvey:
        eventData?.suppress_survey ?? user.ChapterID !== SF_BAY_CHAPTER_ID,
      attendees:
        eventData?.attendees && eventData.attendees.length > 0
          ? [
              ...eventData.attendees.map((name) => ({ name })),
              ...Array(MIN_EMPTY_FIELDS)
                .fill(null)
                .map(() => ({ name: '' })),
            ]
          : Array(DEFAULT_FIELD_COUNT)
              .fill(null)
              .map(() => ({ name: '' })),
      // Scheduled-event fields. Public defaults off, so a new event starts as the
      // quick attendance form; checking "Public event" reveals the rest.
      isPublic: eventData?.is_public ?? false,
      isOnline: eventData?.is_online ?? false,
      description: eventData?.description ?? '',
      // MySQL returns TIME as "HH:MM:SS"; slice to the "HH:MM" the input expects.
      startTime: (eventData?.start_time ?? '').slice(0, 5),
      endTime: (eventData?.end_time ?? '').slice(0, 5),
      timezone: eventData?.timezone || browserTz,
      googlePlaceId: eventData?.location?.google_place_id ?? '',
      locationName: eventData?.location?.name ?? '',
      formattedAddress: eventData?.location?.formatted_address ?? '',
      lat: eventData?.location?.lat ?? undefined,
      lng: eventData?.location?.lng ?? undefined,
    }
  }, [eventData, isConnection, user.ChapterID, mode, browserTz])

  const form = useForm({
    defaultValues: initialValues,
    validators: {
      onSubmit: formSchema,
    },
    onSubmitInvalid: () => {
      toast.error('Please fix the errors before saving')
    },
    onSubmit: async ({ value }) => {
      // Filter out empty attendees.
      const attendeeNames = value.attendees
        .map((a) => (a.name || '').trim())
        .filter((n) => n !== '')

      // Public events carry the richer scheduled-event payload (time, location,
      // description); quick attendance entry and coaching stay the plain payload.
      // Pre-existing events with this data keep it editable either way.
      const includeUpcoming =
        !isConnection && (value.isPublic || editingHasUpcomingData)

      // Attendees are required for quick attendance entry and coaching, but not
      // for public events (those are usually filled in later). Enforced here
      // rather than in a zod refinement because it depends on the trimmed/
      // filtered attendee names and on editingHasUpcomingData, neither of which
      // is a raw form field the schema can see.
      if (!includeUpcoming && attendeeNames.length === 0) {
        toast.error('At least one attendee is required')
        return
      }

      // Check for duplicates.
      const uniqueNames = new Set(attendeeNames)
      if (uniqueNames.size !== attendeeNames.length) {
        toast.error('Please remove duplicates before saving')
        return
      }

      // Calculate diff from original.
      const oldAttendeesSet = new Set(
        (form.options.defaultValues?.attendees || [])
          .map((a) => (a.name || '').trim())
          .filter((n) => n !== ''),
      )

      const addedAttendees = attendeeNames.filter(
        (name) => !oldAttendeesSet.has(name),
      )
      const deletedAttendees = Array.from(oldAttendeesSet).filter(
        (name) => !uniqueNames.has(name),
      )

      await saveEventMutation.mutateAsync({
        event_id: Number(eventId || '0'),
        event_name: value.eventName.trim(),
        event_date: value.eventDate,
        event_type: value.eventType,
        added_attendees: addedAttendees,
        deleted_attendees: deletedAttendees,
        suppress_survey: value.suppressSurvey,
        // Only send scheduled-event fields when they're in play, keeping the
        // plain attendance/connection save payload unchanged.
        ...(includeUpcoming && {
          is_public: value.isPublic,
          is_online: value.isOnline,
          description: value.description.trim(),
          start_time: value.startTime,
          end_time: value.endTime,
          timezone: value.timezone,
          // Online events (or no place picked) have no physical location.
          ...(value.isOnline || !value.googlePlaceId
            ? {}
            : {
                location: {
                  google_place_id: value.googlePlaceId,
                  name: value.locationName,
                  formatted_address: value.formattedAddress,
                  lat: value.lat,
                  lng: value.lng,
                },
              }),
        }),
      })
    },
  })

  const checkForDuplicate = (value: string, currentIndex: number): boolean => {
    const currentAttendees = form.state.values.attendees
    const matches = currentAttendees.filter(
      (a, idx) => idx !== currentIndex && a.name === value,
    )
    return matches.length > 0
  }

  // Subscribe to form state to reactively show/hide the survey checkbox.
  const eventType = useStore(form.store, (state) => state.values.eventType)
  const eventName = useStore(form.store, (state) => state.values.eventName)
  const isDirty = useStore(form.store, (state) => state.isDirty)
  const isOnline = useStore(form.store, (state) => state.values.isOnline)
  const isPublic = useStore(form.store, (state) => state.values.isPublic)
  const eventDate = useStore(form.store, (state) => state.values.eventDate)
  const timezone = useStore(form.store, (state) => state.values.timezone)
  const locationName = useStore(
    form.store,
    (state) => state.values.locationName,
  )

  // The "Public event" checkbox reveals the scheduled-event fields (time,
  // timezone, location, description). Editing an event that already carries
  // that data keeps it visible regardless. Never shown for coaching.
  const showUpcomingFields =
    !isConnection && (isPublic || editingHasUpcomingData)
  // A brand-new public event must be saved before attendance can be recorded —
  // attendance happens later, at the event. Everywhere else (quick attendance
  // entry, coaching, and any saved event) the attendee fields show right away.
  const isNewPublicEvent = !eventId && showUpcomingFields
  const showAttendeeSection = !isNewPublicEvent

  // Predicts whether the server will send a survey by default.
  const shouldShowSuppressSurveyCheckbox = useMemo(() => {
    if (user.ChapterID !== SF_BAY_CHAPTER_ID) return false

    const surveyMatchers = [
      // Surveys are sent for events containing these strings in the name.
      { nameContains: 'chapter meeting' },
      // Surveys are sent for all events of these types.
      { type: 'Action' },
      { type: 'Campaign Action' },
      { type: 'Community' },
      { type: 'Animal Care' },
    ]

    return surveyMatchers.some((matcher) => {
      // If event name, check if included (case-insensitive).
      if (
        matcher.nameContains &&
        !eventName.toLowerCase().includes(matcher.nameContains)
      ) {
        return false
      }
      // If event type, check for match.
      if (matcher.type && matcher.type !== eventType) {
        return false
      }
      return true
    })
  }, [eventName, eventType, user.ChapterID])

  // Subscribe to attendees changes to reactively update the count
  const attendees = useStore(form.store, (state) => state.values.attendees)
  const attendeeCount = useMemo(
    () => attendees.filter((a) => (a.name || '').trim() !== '').length,
    [attendees],
  )

  const ensureMinimumEmptyFields = () => {
    const currentAttendees = form.state.values.attendees
    const emptyCount = currentAttendees.filter((it) => !it.name.length).length
    if (emptyCount < MIN_EMPTY_FIELDS) {
      form.pushFieldValue('attendees', { name: '' })
    }
  }

  // Warn before leaving with unsaved changes.
  useEffect(() => {
    const handleBeforeUnload = (e: BeforeUnloadEvent) => {
      if (isDirty) {
        // `preventDefault()` + setting `returnValue` triggers the
        // browser's native unsaved changes warning dialog. Modern
        // browsers ignore custom messages in returnValue for security,
        // so we use empty string.
        e.preventDefault()
        e.returnValue = ''
      }
    }

    window.addEventListener('beforeunload', handleBeforeUnload)
    return () => window.removeEventListener('beforeunload', handleBeforeUnload)
  }, [isDirty])

  const setDateToToday = () => {
    // "Today" in the event's own timezone, not the browser's, so it lands on
    // the right calendar day around midnight when the two zones differ.
    form.setFieldValue(
      'eventDate',
      todayInTimezone(form.state.values.timezone || browserTz),
    )
  }

  if (eventId && isEventError) {
    return (
      <div className="rounded-md border p-4 text-sm text-destructive">
        Failed to load {isConnection ? 'connection' : 'event'} data.
      </div>
    )
  }

  if (isLoadingActivists || (eventId && isEventLoading) || !activistRegistry) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading...</p>
        </div>
      </div>
    )
  }

  return (
    <form
      key={eventData ? 'loaded' : 'new'}
      onSubmit={async (e) => {
        e.preventDefault()
        e.stopPropagation()
        await form.handleSubmit()
      }}
      className="flex flex-col gap-4"
    >
      {/* Summary toggle. Only when editing a saved event: the detail fields
          collapse behind this bar so attendees sit near the top. The bar itself
          is the toggle in both states (chevron flips), so there's no separate
          collapse button. */}
      {eventId && (
        <button
          type="button"
          onClick={() => setDetailsExpanded((v) => !v)}
          aria-expanded={detailsExpanded}
          className="flex w-full items-center justify-between gap-3 rounded-lg border border-blue-200 bg-blue-50 px-4 py-3 text-left shadow-sm transition-colors hover:bg-blue-100"
        >
          <div className="min-w-0">
            <p className="text-[11px] font-medium uppercase tracking-wide text-muted-foreground">
              {isConnection ? 'Connection' : 'Event'} details
            </p>
            <p className="truncate text-base font-semibold text-foreground">
              {eventName || (isConnection ? 'Connection' : 'Event')}
            </p>
            <p className="truncate text-sm text-muted-foreground">
              {[
                !isConnection && eventType,
                eventDate && format(parseISO(eventDate), 'PPP'),
              ]
                .filter(Boolean)
                .join(' · ')}
            </p>
          </div>
          {detailsExpanded ? (
            <ChevronUp className="h-5 w-5 shrink-0 text-muted-foreground" />
          ) : (
            <ChevronDown className="h-5 w-5 shrink-0 text-muted-foreground" />
          )}
        </button>
      )}

      {/* Detail fields. Always shown for new events; collapsible when editing. */}
      {(!eventId || detailsExpanded) && (
        <>
          {/* Event/Connection Name Field */}
          <form.Field name="eventName">
            {(field) => (
              <div className="flex flex-col gap-2">
                <Label htmlFor="eventName">
                  {isConnection ? 'Coach name' : 'Event name'}
                </Label>
                <Input
                  id="eventName"
                  value={field.state.value ?? ''}
                  onChange={(e) => field.handleChange(e.target.value)}
                  onBlur={field.handleBlur}
                  placeholder={`Enter ${
                    isConnection ? 'connection' : 'event'
                  } name`}
                  className={cn(field.state.meta.errors[0] && 'border-red-500')}
                />
                {field.state.meta.errors[0] && (
                  <p className="text-sm text-red-500">
                    {field.state.meta.errors[0]?.message}
                  </p>
                )}
              </div>
            )}
          </form.Field>

          {/* Event Type Field - Only show for events, not connections */}
          {!isConnection && (
            <form.Field name="eventType">
              {(field) => (
                <div className="flex flex-col gap-2">
                  <Label htmlFor="eventType">Type</Label>
                  <Select
                    value={field.state.value}
                    onValueChange={(value) => field.handleChange(value)}
                  >
                    <SelectTrigger
                      id="eventType"
                      className={cn(
                        field.state.meta.errors[0] && 'border-red-500',
                      )}
                    >
                      <SelectValue placeholder="Select event type" />
                    </SelectTrigger>
                    <SelectContent>
                      {EVENT_TYPES.map((type) => (
                        <SelectItem key={type} value={type}>
                          {type}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  {field.state.meta.errors[0] && (
                    <p className="text-sm text-red-500">
                      {field.state.meta.errors[0]?.message}
                    </p>
                  )}
                </div>
              )}
            </form.Field>
          )}

          {/* Event Date Field */}
          <form.Field name="eventDate">
            {(field) => (
              <div className="flex flex-col gap-2">
                <Label htmlFor="eventDate">Date</Label>
                <div className="flex gap-2">
                  <div className="flex-1">
                    <DatePicker
                      value={
                        field.state.value
                          ? parseISO(field.state.value)
                          : undefined
                      }
                      onValueChange={(date) => {
                        field.handleChange(
                          date ? format(date, 'yyyy-MM-dd') : '',
                        )
                      }}
                      placeholder="Pick a date"
                      className={cn(
                        field.state.meta.errors[0] && 'border-red-500',
                      )}
                    />
                    {field.state.meta.errors[0] && (
                      <p className="text-sm text-red-500 mt-1">
                        {field.state.meta.errors[0]?.message}
                      </p>
                    )}
                  </div>
                  <Button
                    type="button"
                    variant="outline"
                    onClick={setDateToToday}
                  >
                    Today
                  </Button>
                </div>
              </div>
            )}
          </form.Field>

          {/* Public event toggle. Checking it reveals the scheduled-event fields
          below. Hidden for coaching. */}
          {!isConnection && (
            <form.Field name="isPublic">
              {(field) => (
                <div className="flex items-center gap-2">
                  <Checkbox
                    id="isPublic"
                    checked={field.state.value}
                    onCheckedChange={(checked) =>
                      field.handleChange(Boolean(checked))
                    }
                  />
                  <Label htmlFor="isPublic" className="cursor-pointer">
                    Publicly listed event
                  </Label>
                </div>
              )}
            </form.Field>
          )}

          {/* Scheduled-event fields: time, timezone, location, description */}
          {showUpcomingFields && (
            <>
              {/* Start / End Time */}
              <div className="flex gap-4">
                <form.Field name="startTime">
                  {(field) => (
                    <div className="flex flex-1 flex-col gap-2">
                      <Label htmlFor="startTime">
                        Start time{isPublic ? '' : ' (optional)'}
                      </Label>
                      <TimeField
                        aria-label="Start time"
                        value={field.state.value ?? ''}
                        onChange={(v) => field.handleChange(v)}
                        onClear={() => field.handleChange('')}
                        hasError={Boolean(field.state.meta.errors[0])}
                      />
                      {field.state.meta.errors[0] && (
                        <p className="text-sm text-red-500">
                          {field.state.meta.errors[0]?.message}
                        </p>
                      )}
                    </div>
                  )}
                </form.Field>
                <form.Field name="endTime">
                  {(field) => (
                    <div className="flex flex-1 flex-col gap-2">
                      <Label htmlFor="endTime">End time (optional)</Label>
                      <TimeField
                        aria-label="End time"
                        value={field.state.value ?? ''}
                        onChange={(v) => field.handleChange(v)}
                        onClear={() => field.handleChange('')}
                        hasError={Boolean(field.state.meta.errors[0])}
                      />
                      {field.state.meta.errors[0] && (
                        <p className="text-sm text-red-500">
                          {field.state.meta.errors[0]?.message}
                        </p>
                      )}
                    </div>
                  )}
                </form.Field>
              </div>

              {/* Timezone */}
              <form.Field name="timezone">
                {(field) => (
                  <div className="flex flex-col gap-2">
                    <Label htmlFor="timezone">Timezone</Label>
                    <Select
                      value={field.state.value}
                      onValueChange={(value) => field.handleChange(value)}
                    >
                      <SelectTrigger id="timezone">
                        <SelectValue placeholder="Select timezone" />
                      </SelectTrigger>
                      <SelectContent className="max-h-72">
                        {timezones.map((tz) => (
                          <SelectItem key={tz} value={tz}>
                            {tz.replace(/_/g, ' ')}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <p className="text-xs text-muted-foreground">
                      Times are in {timezone.replace(/_/g, ' ')}
                      {getZoneAbbreviation(eventDate, timezone)
                        ? ` (${getZoneAbbreviation(eventDate, timezone)})`
                        : ''}
                      .
                    </p>
                  </div>
                )}
              </form.Field>

              {/* Online checkbox */}
              <form.Field name="isOnline">
                {(field) => (
                  <div className="flex items-center gap-2">
                    <Checkbox
                      id="isOnline"
                      checked={field.state.value}
                      onCheckedChange={(checked) => {
                        const online = Boolean(checked)
                        field.handleChange(online)
                        // Online events have no physical location: clear it.
                        if (online) {
                          form.setFieldValue('googlePlaceId', '')
                          form.setFieldValue('locationName', '')
                          form.setFieldValue('formattedAddress', '')
                          form.setFieldValue('lat', undefined)
                          form.setFieldValue('lng', undefined)
                        }
                      }}
                    />
                    <Label htmlFor="isOnline" className="cursor-pointer">
                      Online event (no physical location)
                    </Label>
                  </div>
                )}
              </form.Field>

              {/* Location (Google Places autocomplete; no free-text) */}
              {!isOnline && (
                <form.Field name="formattedAddress">
                  {(field) => (
                    <div className="flex flex-col gap-2">
                      <Label htmlFor="location">
                        Location{isPublic ? '' : ' (optional)'}
                      </Label>
                      <PlacesAutocomplete
                        id="location"
                        apiKey={googlePlacesApiKey}
                        value={field.state.value ?? ''}
                        locationName={locationName}
                        onSelect={(place) => {
                          form.setFieldValue(
                            'googlePlaceId',
                            place.google_place_id,
                          )
                          form.setFieldValue(
                            'locationName',
                            place.location_name,
                          )
                          form.setFieldValue(
                            'formattedAddress',
                            place.formatted_address,
                          )
                          form.setFieldValue('lat', place.lat)
                          form.setFieldValue('lng', place.lng)
                        }}
                        onClear={() => {
                          form.setFieldValue('googlePlaceId', '')
                          form.setFieldValue('locationName', '')
                          form.setFieldValue('formattedAddress', '')
                          form.setFieldValue('lat', undefined)
                          form.setFieldValue('lng', undefined)
                        }}
                      />
                      {field.state.meta.errors[0] && (
                        <p className="text-sm text-red-500">
                          {field.state.meta.errors[0]?.message}
                        </p>
                      )}
                    </div>
                  )}
                </form.Field>
              )}

              {/* Description */}
              <form.Field name="description">
                {(field) => (
                  <div className="flex flex-col gap-2">
                    <Label htmlFor="description">Description</Label>
                    <Textarea
                      id="description"
                      value={field.state.value ?? ''}
                      onChange={(e) => field.handleChange(e.target.value)}
                      onBlur={field.handleBlur}
                      placeholder="Optional event description"
                      rows={6}
                    />
                  </div>
                )}
              </form.Field>
            </>
          )}

          {/* Suppress Survey Checkbox */}
          {shouldShowSuppressSurveyCheckbox && (
            <form.Field name="suppressSurvey">
              {(field) => (
                <div className="flex items-center gap-2">
                  <Checkbox
                    id="suppressSurvey"
                    checked={field.state.value}
                    onCheckedChange={(checked) =>
                      field.handleChange(Boolean(checked))
                    }
                  />
                  {/* TODO: Consider renaming to "Send survey" with box checked by default. */}
                  <Label htmlFor="suppressSurvey" className="cursor-pointer">
                    Don&apos;t send survey
                  </Label>
                </div>
              )}
            </form.Field>
          )}
        </>
      )}

      {/* Divider between the event detail fields and the attendee section, so
          the two read as distinct sections when the details are expanded. */}
      {(!eventId || detailsExpanded) && <hr className="border-border" />}

      {/* Attendees/Coachees Section. For a brand-new public event this is
          replaced by a prompt to save first; attendance is recorded afterward. */}
      {showAttendeeSection ? (
        <form.Field name="attendees" mode="array">
          {(arrayField) => (
            <div className="flex flex-col gap-2">
              <Label>{isConnection ? 'Coachees' : 'Attendees'}</Label>
              <div className="flex flex-col gap-1">
                {arrayField.state.value.map((_, index) => {
                  const isFocused = index === activeInputIndex
                  return (
                    <form.Field key={index} name={`attendees[${index}].name`}>
                      {(field) => (
                        <AttendeeInputField
                          field={field}
                          index={index}
                          isFocused={isFocused}
                          registry={activistRegistry}
                          checkForDuplicate={checkForDuplicate}
                          inputRef={(el) => {
                            inputRefs.current[index] = el
                          }}
                          onFocus={setActiveInputIndex}
                          onAdvanceFocus={() => {
                            if (index < arrayField.state.value.length - 1) {
                              inputRefs.current[index + 1]?.focus()
                            }
                          }}
                          onChange={ensureMinimumEmptyFields}
                        />
                      )}
                    </form.Field>
                  )
                })}
              </div>
            </div>
          )}
        </form.Field>
      ) : (
        <div className="rounded-md border border-dashed p-4 text-sm text-muted-foreground">
          Save the event first — you can record attendance afterward.
        </div>
      )}

      {/* Save Button with Attendee/Coachee Count */}
      <div className="flex justify-between items-center">
        {showAttendeeSection ? (
          <div className="text-center">
            <p className="text-sm text-gray-500">
              Total {isConnection ? 'coachees' : 'attendees'}
            </p>
            <p className="text-2xl font-bold">{attendeeCount}</p>
          </div>
        ) : (
          <div />
        )}
        <div className="flex items-center gap-4">
          <div className="text-sm">
            {isDirty && (
              <span className="text-red-500 font-medium">Unsaved changes</span>
            )}
          </div>
          <Button
            type="submit"
            variant="default"
            disabled={saveEventMutation.isPending}
          >
            <Save className="h-4 w-4" />
            {saveEventMutation.isPending ? 'Saving...' : 'Save'}
          </Button>
        </div>
      </div>
    </form>
  )
}
