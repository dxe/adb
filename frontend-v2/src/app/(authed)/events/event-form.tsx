'use client'

import { format, parseISO } from 'date-fns'
import { Save } from 'lucide-react'
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
import { useAuthedPageContext } from '@/hooks/useAuthedPageContext'
import { DatePicker } from '@/components/ui/date-picker'
import { EVENT_TYPES, type EventFormMode } from './event-form-schema'
import { useEventForm } from './useEventForm'
import {
  AttendeesSection,
  EventDetailsSummaryBar,
  FieldError,
  ScheduledEventFields,
} from './event-form-sections'

export type { EventFormMode }

type EventFormProps = {
  mode: EventFormMode
  // When editing a saved event, start with the detail fields expanded rather
  // than collapsed behind the summary bar. Used by the post-create confirmation
  // page's "Edit event" link (?expanded=1) so the user lands ready to fix
  // details.
  startExpanded?: boolean
}

export const EventForm = ({ mode, startExpanded }: EventFormProps) => {
  const { googlePlacesApiKey } = useAuthedPageContext()
  const {
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
    isDirty,
  } = useEventForm({ mode, startExpanded })

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
        <EventDetailsSummaryBar
          form={form}
          isConnection={isConnection}
          detailsExpanded={detailsExpanded}
          onToggle={() => setDetailsExpanded((v) => !v)}
        />
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
                <FieldError message={field.state.meta.errors[0]?.message} />
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
                  <FieldError message={field.state.meta.errors[0]?.message} />
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
                    <FieldError
                      message={field.state.meta.errors[0]?.message}
                      className="mt-1"
                    />
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

          {/* Scheduled-event fields: time, timezone, location, description.
              Revealed by the "Publicly listed event" checkbox and grouped in a
              nested sub-box. */}
          {showUpcomingFields && (
            <ScheduledEventFields
              form={form}
              googlePlacesApiKey={googlePlacesApiKey}
            />
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
        <AttendeesSection
          form={form}
          isConnection={isConnection}
          activistRegistry={activistRegistry}
          activeInputIndex={activeInputIndex}
          setActiveInputIndex={setActiveInputIndex}
          inputRefs={inputRefs}
          checkForDuplicate={checkForDuplicate}
          ensureMinimumEmptyFields={ensureMinimumEmptyFields}
        />
      ) : (
        <div className="rounded-md border border-dashed p-4 text-sm text-muted-foreground">
          You can record attendance after saving the event.
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
