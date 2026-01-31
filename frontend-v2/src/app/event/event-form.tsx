'use client'

import { useRef, useState, useEffect, useMemo } from 'react'
import { useForm, useStore } from '@tanstack/react-form'
import { z } from 'zod'
import { Input } from '@/components/ui/input'
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
import { DatePicker } from '@/components/ui/date-picker'
import { format, parseISO } from 'date-fns'
import { Save, Calendar } from 'lucide-react'

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

// Get today's date in YYYY-MM-DD format in the browser's local timezone
const getTodayDate = () => {
  const today = new Date()
  const year = today.getFullYear()
  const month = String(today.getMonth() + 1).padStart(2, '0')
  const day = String(today.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

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

  // Initialize activist registry with IndexedDB caching and incremental sync.
  // The registry loads cached data from IndexedDB first, then syncs any
  // new/updated activists from the server in the background.
  const { registry: activistRegistry, isLoading: isLoadingActivists } =
    useActivistRegistry()

  // Fetch existing event/connection, if editing.
  // (Note: This data is prefetched during SSR for edit pages.)
  const { data: eventData } = useQuery({
    queryKey: [API_PATH.EVENT_GET, eventId],
    queryFn: () => apiClient.getEvent(Number(eventId)),
    enabled: !!eventId,
  })

  const saveEventMutation = useMutation({
    mutationFn: apiClient.saveEvent,
    onSuccess: (result, variables) => {
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

      // Use the attendee order from the current form state, not from the
      // server response (result.attendees). The server returns attendees in
      // arbitrary database order, but we want to preserve the user's input order.
      // This matches the behavior of the legacy Vue version.
      const savedAttendeeNames = form.state.values.attendees
        .map((a) => a.name.trim())
        .filter((n) => n !== '')

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
              ...Array(MIN_EMPTY_FIELDS)
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
    () => attendees.filter((a) => a.name.trim() !== '').length,
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
    form.setFieldValue('eventDate', getTodayDate())
  }

  // Only show loading for activist list since event data is prefetched during SSR
  if (isLoadingActivists || !activistRegistry) {
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
      className="flex flex-col gap-4"
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
              <Select
                value={field.state.value}
                onValueChange={(value) => field.handleChange(value)}
              >
                <SelectTrigger
                  id="eventType"
                  className={cn(field.state.meta.errors[0] && 'border-red-500')}
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
                    field.state.value ? parseISO(field.state.value) : undefined
                  }
                  onValueChange={(date) => {
                    field.handleChange(date ? format(date, 'yyyy-MM-dd') : '')
                  }}
                  placeholder="Pick a date"
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
              {/* TODO: Consider renaming to "Send survey" with box checked by default. */}
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
