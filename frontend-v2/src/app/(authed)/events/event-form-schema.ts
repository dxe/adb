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

// Curated location-name suggestions, offered (but not forced) on the location
// field so recurring venues get a single canonical spelling instead of drifting
// into "ARC", "Berkeley ARC", etc. Free text is still allowed.
export const SUGGESTED_LOCATION_NAMES = [
  'Berkeley Animal Rights Center',
] as const

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
    // Location: a free-text display name (always editable) plus optional geo
    // data. googlePlaceId/formattedAddress/lat/lng come from a Google Places
    // pick; lat/lng can also be entered by hand for spots that aren't a clean
    // Place, like an intersection. The name is stored on the event itself, so
    // editing it never affects another event.
    locationName: z.string(),
    googlePlaceId: z.string(),
    formattedAddress: z.string(),
    lat: z.number().optional(),
    lng: z.number().optional(),
    // UI-only: when true, show manual latitude/longitude inputs instead of the
    // Google Places search. Not submitted to the server.
    manualLocation: z.boolean(),
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
  // In-person public events need a location name; online ones don't. Geo data
  // (a Google place or coordinates) is optional — the name is what's required.
  .refine((v) => !v.isPublic || v.isOnline || v.locationName.trim() !== '', {
    message: 'Location is required for in-person public events',
    path: ['locationName'],
  })
  // Manual coordinates, when provided, must be valid.
  .refine((v) => v.lat === undefined || (v.lat >= -90 && v.lat <= 90), {
    message: 'Latitude must be between -90 and 90',
    path: ['lat'],
  })
  .refine((v) => v.lng === undefined || (v.lng >= -180 && v.lng <= 180), {
    message: 'Longitude must be between -180 and 180',
    path: ['lng'],
  })

export type FormValues = z.infer<typeof formSchema>
