'use client'

import { useStore } from '@tanstack/react-form'
import { format, parseISO } from 'date-fns'
import { ChevronDown, ChevronUp } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Checkbox } from '@/components/ui/checkbox'
import { Textarea } from '@/components/ui/textarea'
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

// Online events (or no place picked) have no physical location: clear it. Used
// by both the online-checkbox handler and the location field's onClear.
export const clearLocation = (form: EventFormApi) => {
  form.setFieldValue('googlePlaceId', '')
  form.setFieldValue('locationName', '')
  form.setFieldValue('formattedAddress', '')
  form.setFieldValue('lat', undefined)
  form.setFieldValue('lng', undefined)
}

export const EventDetailsSummaryBar = ({
  form,
  isConnection,
  detailsExpanded,
  onToggle,
}: {
  form: EventFormApi
  isConnection: boolean
  detailsExpanded: boolean
  onToggle: () => void
}) => {
  const eventName = useStore(form.store, (state) => state.values.eventName)
  const eventType = useStore(form.store, (state) => state.values.eventType)
  const eventDate = useStore(form.store, (state) => state.values.eventDate)

  return (
    <button
      type="button"
      onClick={onToggle}
      aria-expanded={detailsExpanded}
      className="flex w-full items-center justify-between gap-3 rounded-lg border border-blue-200 bg-blue-50 px-4 py-3 text-left shadow-sm transition-colors hover:bg-blue-100"
    >
      <div className="min-w-0">
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
  )
}

export const ScheduledEventFields = ({
  form,
  googlePlacesApiKey,
}: {
  form: EventFormApi
  googlePlacesApiKey: string
}) => {
  const isPublic = useStore(form.store, (state) => state.values.isPublic)
  const isOnline = useStore(form.store, (state) => state.values.isOnline)
  const eventDate = useStore(form.store, (state) => state.values.eventDate)
  const locationName = useStore(
    form.store,
    (state) => state.values.locationName,
  )
  const manualLocation = useStore(
    form.store,
    (state) => state.values.manualLocation,
  )
  const timezones = getCommonTimezones()

  return (
    <div className="flex flex-col gap-4 rounded-lg border bg-muted/30 p-4">
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

      {/* Location: Google Places autocomplete, or a manual free-text entry for
          spots that aren't a clean Place (intersections, public land, etc.). */}
      {!isOnline && (
        <div className="flex flex-col gap-2">
          <Label htmlFor="location">
            Location{isPublic ? '' : ' (optional)'}
          </Label>
          {manualLocation ? (
            <>
              <form.Field name="locationName">
                {(field) => (
                  <div className="flex flex-col gap-2">
                    <Input
                      id="location"
                      value={field.state.value ?? ''}
                      onChange={(e) => field.handleChange(e.target.value)}
                      onBlur={field.handleBlur}
                      placeholder="e.g. 16th & Mission St (NW corner)"
                      className={cn(
                        field.state.meta.errors[0] && 'border-red-500',
                      )}
                    />
                    <FieldError message={field.state.meta.errors[0]?.message} />
                  </div>
                )}
              </form.Field>
              <div className="flex gap-4">
                <form.Field name="lat">
                  {(field) => (
                    <div className="flex flex-1 flex-col gap-2">
                      <Label htmlFor="lat">Latitude (optional)</Label>
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
                        placeholder="37.7749"
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
                      <Label htmlFor="lng">Longitude (optional)</Label>
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
                        placeholder="-122.4194"
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
              <button
                type="button"
                className="self-start text-sm text-primary hover:underline"
                onClick={() => {
                  clearLocation(form)
                  form.setFieldValue('manualLocation', false)
                }}
              >
                Search Google Places instead
              </button>
            </>
          ) : (
            <>
              <form.Field name="formattedAddress">
                {(field) => (
                  <div className="flex flex-col gap-2">
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
                        form.setFieldValue('locationName', place.location_name)
                        form.setFieldValue(
                          'formattedAddress',
                          place.formatted_address,
                        )
                        form.setFieldValue('lat', place.lat)
                        form.setFieldValue('lng', place.lng)
                      }}
                      onClear={() => clearLocation(form)}
                    />
                    <FieldError message={field.state.meta.errors[0]?.message} />
                  </div>
                )}
              </form.Field>
              <button
                type="button"
                className="self-start text-sm text-primary hover:underline"
                onClick={() => {
                  // Switching to manual: a free-text entry isn't a Google place.
                  form.setFieldValue('googlePlaceId', '')
                  form.setFieldValue('manualLocation', true)
                }}
              >
                Enter location manually
              </button>
            </>
          )}
        </div>
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
