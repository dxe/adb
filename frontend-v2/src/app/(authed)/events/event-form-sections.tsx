'use client'

import { useStore } from '@tanstack/react-form'
import { format, parseISO } from 'date-fns'
import { ChevronDown, ChevronLeft } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Checkbox } from '@/components/ui/checkbox'
import { Textarea } from '@/components/ui/textarea'
import { Collapse } from '@/components/ui/collapse'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { TimeField } from '@/components/ui/time-field'
import { PlacesAutocomplete } from './places-autocomplete'
import { AttendeeInputField } from './attendee-input-field'
import { getCommonTimezones, getZoneAbbreviation } from '@/lib/time'
import type { ActivistRegistry } from './activist-registry'
import type { EventFormApi } from './useEventForm'

// Replaces the repeated field-error paragraph. The date field renders this with
// an extra `mt-1`, so callers can opt into that spacing.
export const FieldError = ({
  message,
  className,
}: {
  message?: string
  className?: string
}) => {
  if (!message) return null
  return (
    <p className={`text-sm text-red-500${className ? ` ${className}` : ''}`}>
      {message}
    </p>
  )
}

// Muted sub-heading that separates the scheduled-event fields into groups
// (When / Where / Details) so each block reads as a distinct section.
const SectionHeading = ({ children }: { children: React.ReactNode }) => (
  <h3 className="text-xs font-semibold uppercase tracking-wide text-blue-900/60">
    {children}
  </h3>
)

// Clears the geo data attached to a location (Google place + coordinates) while
// leaving the free-text name intact. Used when the user edits the place search
// away from a selection, or switches between the search and manual coordinates.
export const clearGeo = (form: EventFormApi) => {
  form.setFieldValue('googlePlaceId', '')
  form.setFieldValue('formattedAddress', '')
  form.setFieldValue('lat', undefined)
  form.setFieldValue('lng', undefined)
}

// Online events have no physical location: clear the name and geo. Used by the
// online-checkbox handler.
export const clearLocation = (form: EventFormApi) => {
  form.setFieldValue('locationName', '')
  clearGeo(form)
}

// Collapsible card wrapping the event detail fields when editing a saved event.
// Collapsed, the header shows a summary (name · type · date) so attendees sit
// near the top; expanded, the detail fields render *inside* the same bordered
// container (rather than below it) so the form reads as one contained unit.
export const EventDetailsCard = ({
  form,
  isConnection,
  detailsExpanded,
  onToggle,
  children,
}: {
  form: EventFormApi
  isConnection: boolean
  detailsExpanded: boolean
  onToggle: () => void
  children: React.ReactNode
}) => {
  const eventName = useStore(form.store, (state) => state.values.eventName)
  const eventType = useStore(form.store, (state) => state.values.eventType)
  const eventDate = useStore(form.store, (state) => state.values.eventDate)

  return (
    <div className="overflow-hidden rounded-lg border border-blue-200 bg-blue-50 shadow-sm">
      <button
        type="button"
        onClick={onToggle}
        aria-expanded={detailsExpanded}
        className="flex w-full items-center justify-between gap-3 px-4 py-3 text-left transition-colors hover:bg-blue-100"
      >
        <div className="min-w-0">
          {detailsExpanded ? (
            <p className="text-base font-semibold text-foreground">Details</p>
          ) : (
            <>
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
            </>
          )}
        </div>
        {detailsExpanded ? (
          <ChevronDown className="h-5 w-5 shrink-0 text-muted-foreground" />
        ) : (
          <ChevronLeft className="h-5 w-5 shrink-0 text-muted-foreground" />
        )}
      </button>
      <Collapse open={detailsExpanded}>
        <div className="flex flex-col gap-4 border-t border-blue-200 bg-background p-4">
          {children}
        </div>
      </Collapse>
    </div>
  )
}

export const ScheduledEventFields = ({
  form,
  googlePlacesApiKey,
  // Whether this section is currently revealed. Gates the Google Maps load so it
  // stays dormant while the (still-mounted, for animation) section is collapsed.
  active = true,
}: {
  form: EventFormApi
  googlePlacesApiKey: string
  active?: boolean
}) => {
  const isPublic = useStore(form.store, (state) => state.values.isPublic)
  const isOnline = useStore(form.store, (state) => state.values.isOnline)
  const eventDate = useStore(form.store, (state) => state.values.eventDate)
  const manualLocation = useStore(
    form.store,
    (state) => state.values.manualLocation,
  )
  const timezones = getCommonTimezones()

  return (
    <div className="flex flex-col gap-6 rounded-lg border border-blue-200 bg-blue-50/40 p-4">
      {/* When: start/end time + timezone */}
      <div className="flex flex-col gap-4">
        <SectionHeading>When</SectionHeading>
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
                <FieldError message={field.state.meta.errors[0]?.message} />
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
                <FieldError message={field.state.meta.errors[0]?.message} />
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
                  {timezones.map((tz) => {
                    const abbr = getZoneAbbreviation(eventDate, tz)
                    return (
                      <SelectItem key={tz} value={tz}>
                        {`${tz.replace(/_/g, ' ')}${abbr ? ` (${abbr})` : ''}`}
                      </SelectItem>
                    )
                  })}
                </SelectContent>
              </Select>
            </div>
          )}
        </form.Field>
      </div>

      {/* Where: online toggle + physical location */}
      <div className="flex flex-col gap-4">
        <SectionHeading>Where</SectionHeading>
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
                  if (online) {
                    clearLocation(form)
                  }
                }}
              />
              <Label htmlFor="isOnline" className="cursor-pointer">
                Online event (no physical location)
              </Label>
            </div>
          )}
        </form.Field>

        {/* Location: pick a place (or enter coordinates), then confirm/edit the
          display name filled in from it. Online events have no location. */}
        <Collapse open={!isOnline} className={cn(isOnline && '-mt-4')}>
          <div className="flex flex-col gap-3">
            {/* Geo first: a Google Places search (default) or manual coordinates
              for spots that aren't a clean Place (intersections, public land). */}
            {manualLocation ? (
              <div className="flex gap-4">
                <form.Field name="lat">
                  {(field) => (
                    <div className="flex flex-1 flex-col gap-2">
                      <Label htmlFor="lat">
                        Latitude{isPublic ? '' : ' (optional)'}
                      </Label>
                      <Input
                        id="lat"
                        type="number"
                        inputMode="decimal"
                        step="any"
                        value={field.state.value ?? ''}
                        onChange={(e) => {
                          const n = e.target.valueAsNumber
                          field.handleChange(Number.isNaN(n) ? undefined : n)
                        }}
                        placeholder="e.g. 37.7749"
                        className={cn(
                          field.state.meta.errors[0] && 'border-red-500',
                        )}
                      />
                      <FieldError
                        message={field.state.meta.errors[0]?.message}
                      />
                    </div>
                  )}
                </form.Field>
                <form.Field name="lng">
                  {(field) => (
                    <div className="flex flex-1 flex-col gap-2">
                      <Label htmlFor="lng">
                        Longitude{isPublic ? '' : ' (optional)'}
                      </Label>
                      <Input
                        id="lng"
                        type="number"
                        inputMode="decimal"
                        step="any"
                        value={field.state.value ?? ''}
                        onChange={(e) => {
                          const n = e.target.valueAsNumber
                          field.handleChange(Number.isNaN(n) ? undefined : n)
                        }}
                        placeholder="e.g. -122.4194"
                        className={cn(
                          field.state.meta.errors[0] && 'border-red-500',
                        )}
                      />
                      <FieldError
                        message={field.state.meta.errors[0]?.message}
                      />
                    </div>
                  )}
                </form.Field>
              </div>
            ) : (
              <form.Field name="formattedAddress">
                {(field) => (
                  <div className="flex flex-col gap-2">
                    <Label htmlFor="location">
                      Location address{isPublic ? '' : ' (optional)'}
                    </Label>
                    <PlacesAutocomplete
                      id="location"
                      apiKey={googlePlacesApiKey}
                      active={active && !isOnline}
                      value={field.state.value ?? ''}
                      onSelect={(place) => {
                        form.setFieldValue(
                          'googlePlaceId',
                          place.google_place_id,
                        )
                        form.setFieldValue(
                          'formattedAddress',
                          place.formatted_address,
                        )
                        form.setFieldValue('lat', place.lat)
                        form.setFieldValue('lng', place.lng)
                        // Fill the editable display name from the picked place,
                        // but only when Google gave a real name (not a bare
                        // address) — otherwise leave whatever's there for the
                        // user to fill in.
                        if (place.location_name) {
                          form.setFieldValue(
                            'locationName',
                            place.location_name,
                          )
                        }
                      }}
                      onClear={() => clearGeo(form)}
                    />
                    <FieldError message={field.state.meta.errors[0]?.message} />
                  </div>
                )}
              </form.Field>
            )}

            {/* Toggle between the place search and manual coordinates, right under
              the geo control it switches. */}
            <button
              type="button"
              className="-mt-2 self-start text-sm text-primary hover:underline"
              onClick={() => {
                clearGeo(form)
                form.setFieldValue('manualLocation', !manualLocation)
              }}
            >
              {manualLocation
                ? 'Search for a place instead'
                : 'Enter coordinates manually'}
            </button>

            {/* Display name: filled in from the picked place (or typed for a
              coords-only spot), always editable, and stored on the event itself
              — so correcting it never affects another event. */}
            <form.Field name="locationName">
              {(field) => (
                <div className="flex flex-col gap-2">
                  <Label htmlFor="locationName">
                    Location name{isPublic ? '' : ' (optional)'}
                  </Label>
                  <Input
                    id="locationName"
                    value={field.state.value ?? ''}
                    onChange={(e) => field.handleChange(e.target.value)}
                    onBlur={field.handleBlur}
                    placeholder="e.g. Berkeley Animal Rights Center"
                    className={cn(
                      field.state.meta.errors[0] && 'border-red-500',
                    )}
                  />
                  <FieldError message={field.state.meta.errors[0]?.message} />
                </div>
              )}
            </form.Field>
          </div>
        </Collapse>
      </div>

      {/* Information: free-form description */}
      <div className="flex flex-col gap-4">
        <SectionHeading>Information</SectionHeading>
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
      </div>
    </div>
  )
}

export const AttendeesSection = ({
  form,
  isConnection,
  activistRegistry,
  activeInputIndex,
  setActiveInputIndex,
  inputRefs,
  checkForDuplicate,
  ensureMinimumEmptyFields,
}: {
  form: EventFormApi
  isConnection: boolean
  activistRegistry: ActivistRegistry
  activeInputIndex: number
  setActiveInputIndex: (index: number) => void
  inputRefs: React.RefObject<(HTMLInputElement | null)[]>
  checkForDuplicate: (value: string, currentIndex: number) => boolean
  ensureMinimumEmptyFields: () => void
}) => {
  return (
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
  )
}
