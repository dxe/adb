'use client'

import { useRef, useState, useEffect, useMemo } from 'react'
import { useForm, useStore } from '@tanstack/react-form'
import { z } from 'zod'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { Checkbox } from '@/components/ui/checkbox'
import { cn } from '@/lib/utils'
import { API_PATH, apiClient } from '@/lib/api'
import { useQuery, useQueryClient, useMutation } from '@tanstack/react-query'
import { useParams, useRouter } from 'next/navigation'
import toast from 'react-hot-toast'
import { useAuthedPageContext } from '@/hooks/useAuthedPageContext'
import { SF_BAY_CHAPTER_ID } from '@/lib/constants'
import { ActivistRegistry } from '@/lib/activist-registry'
import { AttendeeInputField } from './attendee-input-field'

// TODO(jh):
// - test in prod
// - improve styling
// - replace vue page w/ react page & update api to not return a redirect response on save
// - LATER: store list of names from server in indexed db & only update what's been created, updated, or deleted since last load?
// - LATER: store unsaved data in session storage to prevent accidental loss?
// - LATER: dark mode

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

const formSchema = z.object({
  eventName: z.string().min(1, 'Event name is required'),
  eventType: z.string().min(1, 'Event type is required'),
  eventDate: z.string().min(1, 'Event date is required'),
  suppressSurvey: z.boolean(),
  attendees: z.array(attendeeSchema),
})

type FormValues = z.infer<typeof formSchema>

type EventFormProps = {
  mode: 'event' | 'connection'
}

export const EventForm = ({ mode }: EventFormProps) => {
  const router = useRouter()
  const params = useParams()
  const queryClient = useQueryClient()
  const { user } = useAuthedPageContext()
  const eventId = params.id ? String(params.id) : undefined
  const isConnection = mode === 'connection'

  const inputRefs = useRef<(HTMLInputElement | null)[]>(
    Array(DEFAULT_FIELD_COUNT).fill(null),
  )
  const [activeInputIndex, setActiveInputIndex] = useState(0)

  // Fetch activist list from server.
  // We don't want to do this during SSR b/c it's several MB,
  // and eventually we'd like to cache this data on the client
  // to avoid sending it on every page load.
  const { data: activistData, isLoading: isLoadingActivists } = useQuery({
    queryKey: [API_PATH.ACTIVIST_LIST_BASIC],
    queryFn: apiClient.getActivistListBasic,
  })

  // Fetch existing event/connection, if editing.
  // (Note: This data is prefetched during SSR for edit pages.)
  const { data: eventData } = useQuery({
    queryKey: [API_PATH.EVENT_GET, eventId],
    queryFn: () => apiClient.getEvent(Number(eventId)),
    enabled: !!eventId,
  })

  // Create activist registry for lookups.
  const registry = useMemo(
    () => new ActivistRegistry(activistData?.activists || []),
    [activistData?.activists],
  )

  // Mutation for saving event.
  const saveEventMutation = useMutation({
    mutationFn: apiClient.saveEvent,
    onSuccess: (result, variables) => {
      if (result.status === 'error') {
        toast.error(
          result.message ||
            `Error saving ${isConnection ? 'connection' : 'event'}`,
        )
        return
      }

      toast.success(`${isConnection ? 'Connection' : 'Event'} saved!`)

      // TODO(jh): once the vue page is removed, update the api to just
      // return a json payload w/ the event id instead of a redirect.
      // for now, we'll extract the id from the redirect url to stay
      // in the react app for testing.
      if (result.redirect) {
        // Parse redirect like "/update_event/8" or "/update_connection/5"
        const match = result.redirect.match(
          /\/(update_event|update_connection)\/(\d+)/,
        )
        if (match) {
          const newEventId = match[2]
          // Update URL to include the new event ID.
          const newPath = isConnection
            ? `/coaching/${newEventId}`
            : `/event/${newEventId}`
          router.push(newPath)
        } else {
          // Fallback: redirect to legacy route if we can't parse it.
          window.location.href = result.redirect
          return
        }
      }

      // Reset the form's dirty state after successful save.
      // This prevents "unsaved changes" warning after successful save.
      // TanStack form has a bug requiring the `keepDefaultValues` option:
      // https://github.com/TanStack/form/issues/1798.
      form.reset(
        {
          eventName: variables.event_name,
          eventType: variables.event_type,
          eventDate: variables.event_date,
          suppressSurvey: variables.suppress_survey,
          attendees: (result.attendees ?? [])
            .map((name) => ({ name }))
            .concat(
              Array(MIN_EMPTY_FIELDS)
                .fill(null)
                .map(() => ({ name: '' })),
            ),
        },
        { keepDefaultValues: true },
      )

      // Refresh activist list to include newly created activists.
      // This ensures they appear in autocomplete suggestions.
      queryClient.invalidateQueries({
        queryKey: [API_PATH.ACTIVIST_LIST_BASIC],
      })
      queryClient.invalidateQueries({
        queryKey: [API_PATH.EVENT_GET, eventId],
      })
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Error saving event')
    },
  })

  const initialValues: FormValues = useMemo(() => {
    if (eventId && !eventData) {
      throw new Error('Expected event data to be prefetched')
    }
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
              ...Array(Math.max(0, MIN_EMPTY_FIELDS))
                .fill(null)
                .map(() => ({ name: '' })),
            ]
          : Array(DEFAULT_FIELD_COUNT)
              .fill(null)
              .map(() => ({ name: '' })),
    }
  }, [eventData, eventId, isConnection, user.ChapterID])

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
        .map((a) => a.name.trim())
        .filter((n) => n !== '')

      if (attendeeNames.length === 0) {
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
          .map((a) => a.name.trim())
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
      })
    },
  })

  // Helper functions.
  const checkForDuplicate = (value: string, currentIndex: number): boolean => {
    const attendees = form.state.values.attendees
    const matches = attendees.filter(
      (a, idx) => idx !== currentIndex && a.name === value,
    )
    return matches.length > 0
  }

  // Subscribe to form state to reactively show/hide the survey checkbox.
  const eventType = useStore(form.store, (state) => state.values.eventType)
  const eventName = useStore(form.store, (state) => state.values.eventName)

  const shouldShowSuppressSurveyCheckbox = useMemo(() => {
    // Only ever shown for SF Bay Area chapter.
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
        eventName.toLowerCase().includes(matcher.nameContains)
      ) {
        return true
      }
      // If event type, check for match.
      if (matcher.type && matcher.type === eventType) {
        return true
      }
      return false
    })
  }, [eventName, eventType, user.ChapterID])

  const attendeeCount = useMemo(
    () =>
      form.state.values.attendees.filter((a) => a.name.trim() !== '').length,
    [form.state.values.attendees],
  )

  const ensureMinimumEmptyFields = () => {
    const emptyCount = form.state.values.attendees.filter(
      (it) => !it.name.length,
    ).length
    if (emptyCount < MIN_EMPTY_FIELDS) {
      form.pushFieldValue('attendees', { name: '' })
    }
  }

  // Warn before leaving with unsaved changes.
  useEffect(() => {
    const handleBeforeUnload = (e: BeforeUnloadEvent) => {
      if (form.state.isDirty) {
        e.preventDefault()
        e.returnValue = ''
      }
    }

    window.addEventListener('beforeunload', handleBeforeUnload)
    return () => window.removeEventListener('beforeunload', handleBeforeUnload)
  }, [form.state.isDirty])

  const setDateToToday = () => {
    const today = new Date().toISOString().split('T')[0]
    form.setFieldValue('eventDate', today)
  }

  // Only show loading for activist list since event data is prefetched during SSR
  if (isLoadingActivists) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading activist names...</p>
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
      className="flex flex-col gap-6"
    >
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
              placeholder={`Enter ${isConnection ? 'connection' : 'event'} name`}
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
              <select
                id="eventType"
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                onBlur={field.handleBlur}
                className={cn(
                  'flex h-9 w-full items-center justify-between whitespace-nowrap rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm ring-offset-background focus:outline-none focus:ring-1 focus:ring-ring disabled:cursor-not-allowed disabled:opacity-50',
                  field.state.meta.errors[0] && 'border-red-500',
                )}
              >
                <option value="" disabled>
                  Select event type
                </option>
                {EVENT_TYPES.map((type) => (
                  <option key={type} value={type}>
                    {type}
                  </option>
                ))}
              </select>
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
                <Input
                  id="eventDate"
                  type="date"
                  value={field.state.value}
                  onChange={(e) => field.handleChange(e.target.value)}
                  onBlur={field.handleBlur}
                  className={cn(field.state.meta.errors[0] && 'border-red-500')}
                />
                {field.state.meta.errors[0] && (
                  <p className="text-sm text-red-500 mt-1">
                    {field.state.meta.errors[0]?.message}
                  </p>
                )}
              </div>
              <Button type="button" variant="outline" onClick={setDateToToday}>
                Today
              </Button>
            </div>
          </div>
        )}
      </form.Field>

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
              <Label htmlFor="suppressSurvey" className="cursor-pointer">
                Don&apos;t send survey
              </Label>
            </div>
          )}
        </form.Field>
      )}

      {/* Attendees/Coachees Section */}
      <form.Field name="attendees" mode="array">
        {(arrayField) => (
          <div className="flex flex-col gap-2">
            <Label>{isConnection ? 'Coachees' : 'Attendees'}</Label>
            <div className="flex flex-col gap-4">
              {arrayField.state.value.map((_, index) => {
                const isFocused = index === activeInputIndex
                return (
                  <form.Field key={index} name={`attendees[${index}].name`}>
                    {(field) => (
                      <AttendeeInputField
                        field={field}
                        index={index}
                        isFocused={isFocused}
                        registry={registry}
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
                        ensureMinimumEmptyFields={ensureMinimumEmptyFields}
                      />
                    )}
                  </form.Field>
                )
              })}
            </div>
          </div>
        )}
      </form.Field>

      {/* Save Button with Attendee/Coachee Count */}
      <div className="flex justify-between items-center">
        <div className="text-center">
          <p className="text-sm text-gray-500">
            Total {isConnection ? 'coachees' : 'attendees'}
          </p>
          <p className="text-2xl font-bold">{attendeeCount}</p>
        </div>
        <div className="flex items-center gap-4">
          <div className="text-sm">
            {form.state.isDirty && (
              <span className="text-red-500 font-medium">Unsaved changes</span>
            )}
          </div>
          <Button
            type="submit"
            variant="default"
            disabled={saveEventMutation.isPending}
          >
            {saveEventMutation.isPending ? 'Saving...' : 'Save'}
          </Button>
        </div>
      </div>
    </form>
  )
}
