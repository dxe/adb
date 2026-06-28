'use client'

import { useRef, useState, useEffect, useMemo } from 'react'
import { useForm, useStore } from '@tanstack/react-form'
import { API_PATH, apiClient } from '@/lib/api'
import { useQuery, useQueryClient, useMutation } from '@tanstack/react-query'
import { useParams, useRouter } from 'next/navigation'
import toast from 'react-hot-toast'
import { useAuthedPageContext } from '@/hooks/useAuthedPageContext'
import { SF_BAY_CHAPTER_ID } from '@/lib/constants'
import { useActivistRegistry } from './useActivistRegistry'
import { getBrowserTimezone, todayInTimezone } from '@/lib/time'
import {
  DEFAULT_FIELD_COUNT,
  MIN_EMPTY_FIELDS,
  formSchema,
  type EventFormMode,
  type FormValues,
} from './event-form-schema'

type UseEventFormArgs = {
  mode: EventFormMode
  startExpanded?: boolean
  // For a public event the attendee fields default to collapsed behind an "Add
  // attendees" link (attendance is usually recorded later). Callers that open
  // an event to manage it — the event list, home, "Take attendance now" — pass
  // true so the fields are visible immediately.
  startAttendeesExpanded?: boolean
}

export type EventFormApi = ReturnType<typeof useEventForm>['form']

export const useEventForm = ({
  mode,
  startExpanded,
  startAttendeesExpanded,
}: UseEventFormArgs) => {
  const router = useRouter()
  const params = useParams()
  const queryClient = useQueryClient()
  const { user } = useAuthedPageContext()
  const eventId = params.id ? String(params.id) : undefined
  const isConnection = mode === 'connection'
  const browserTz = useMemo(() => getBrowserTimezone(), [])

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
      eventData.location?.name ||
      eventData.description),
  )

  // When editing a saved event the detail fields (name, type, date, schedule,
  // location, description) are usually already set, so they collapse behind a
  // summary header to keep the attendee list near the top for taking
  // attendance. New events start expanded since you're filling them out.
  const [detailsExpanded, setDetailsExpanded] = useState(
    startExpanded ?? !eventId,
  )
  // A public event hides the attendee fields behind an "Add attendees" link
  // (attendance is usually taken later, at the event), but lets the user open
  // them to record attendees up front. Defaults to collapsed; callers that open
  // an event to manage it pass startAttendeesExpanded to show them outright.
  const [attendeesExpanded, setAttendeesExpanded] = useState(
    startAttendeesExpanded ?? false,
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
        manualLocation: current.manualLocation,
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
      startTime: eventData?.start_time ?? '',
      endTime: eventData?.end_time ?? '',
      timezone: eventData?.timezone || browserTz,
      locationName: eventData?.location?.name ?? '',
      googlePlaceId: eventData?.location?.google_place_id ?? '',
      formattedAddress: eventData?.location?.formatted_address ?? '',
      lat: eventData?.location?.lat ?? undefined,
      lng: eventData?.location?.lng ?? undefined,
      // Show the manual-coordinates inputs (instead of the Google search) when a
      // saved location has coordinates but isn't a Google place.
      manualLocation: Boolean(
        eventData?.location &&
        !eventData.location.google_place_id &&
        (eventData.location.lat != null || eventData.location.lng != null),
      ),
    }
  }, [eventData, isConnection, user.ChapterID, browserTz])

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

      // Resolve the location payload: a free-text name plus optional geo data (a
      // Google place and/or coordinates). Nothing is sent for an online event or
      // when no location was given.
      const locationPayload = (() => {
        if (value.isOnline) return undefined
        const name = value.locationName.trim()
        const address = value.formattedAddress.trim()
        const lat = Number.isFinite(value.lat) ? value.lat : undefined
        const lng = Number.isFinite(value.lng) ? value.lng : undefined
        const hasGeo =
          Boolean(value.googlePlaceId) ||
          Boolean(address) ||
          lat !== undefined ||
          lng !== undefined
        if (!name && !hasGeo) return undefined
        return {
          google_place_id: value.googlePlaceId,
          name,
          formatted_address: address || name,
          lat,
          lng,
        }
      })()

      await saveEventMutation.mutateAsync({
        event_id: Number(eventId || '0'),
        event_name: value.eventName.trim(),
        event_date: value.eventDate,
        event_type: value.eventType,
        added_attendees: addedAttendees,
        deleted_attendees: deletedAttendees,
        suppress_survey: value.suppressSurvey,
        ...(!isConnection && {
          is_public: value.isPublic,
          is_online: value.isOnline,
          description: value.description.trim(),
          start_time: value.startTime,
          end_time: value.endTime,
          timezone: value.timezone,
          ...(locationPayload ? { location: locationPayload } : {}),
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
  const isPublic = useStore(form.store, (state) => state.values.isPublic)

  // The "Public event" checkbox reveals the scheduled-event fields (time,
  // timezone, location, description). Editing an event that already carries
  // that data keeps it visible regardless. Never shown for coaching.
  const showUpcomingFields =
    !isConnection && (isPublic || editingHasUpcomingData)
  // For a public event, attendance is usually recorded later (at the event), so
  // the attendee fields stay tucked behind an "Add attendees" link until they're
  // needed — whether the event is brand new or already saved. Everywhere else
  // (quick attendance entry, coaching) the fields show right away.
  const showAttendeeSection = !showUpcomingFields

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
    // Once the user has started entering attendees, keep the section open even
    // if the names are later cleared, so it doesn't collapse out from under them
    // mid-edit (e.g. after checking "Public event" on a form that already has
    // attendees).
    setAttendeesExpanded(true)
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

  return {
    form,
    eventId,
    eventData,
    isConnection,
    isEventLoading,
    isEventError,
    isLoadingActivists,
    activistRegistry,
    detailsExpanded,
    setDetailsExpanded,
    attendeesExpanded,
    setAttendeesExpanded,
    inputRefs,
    activeInputIndex,
    setActiveInputIndex,
    checkForDuplicate,
    ensureMinimumEmptyFields,
    setDateToToday,
    saveEventMutation,
    attendeeCount,
    shouldShowSuppressSurveyCheckbox,
    showUpcomingFields,
    showAttendeeSection,
    editingHasUpcomingData,
    isDirty,
    isPublic,
  }
}
