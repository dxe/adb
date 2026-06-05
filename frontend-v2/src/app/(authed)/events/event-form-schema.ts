import { z } from 'zod'

export const EVENT_TYPES = [
  'Action',
  'Campaign Action',
  'Community',
  'Frontline Surveillance',
  'Meeting',
  'Outreach',
  'Animal Care',
  'Training',
] as const

export const DEFAULT_FIELD_COUNT = 5
export const MIN_EMPTY_FIELDS = 1

export type EventFormMode = 'event' | 'connection'

// Zod schema for form validation.
export const attendeeSchema = z.object({
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

export const formSchema = z
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

export type FormValues = z.infer<typeof formSchema>
