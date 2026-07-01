'use client'

import { format, parseISO } from 'date-fns'
import { Info, Save } from 'lucide-react'
import { Input } from '@/components/ui/input'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
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
  EventDetailsCard,
  ScheduledEventFields,
} from './event-form-sections'
import { Switch } from '@/components/ui/switch'
import { Collapse } from '@/components/ui/collapse'
import { FieldError } from './field-error'

export type { EventFormMode }

type EventFormProps = {
  mode: EventFormMode
  // When editing a saved event, start with the detail fields expanded rather
  // than collapsed behind the summary bar. Used by the post-create confirmation
  // page's "Edit event" link (?expanded=1) so the user lands ready to fix
  // details.
  startExpanded?: boolean
  // Show the attendee fields immediately rather than behind the "Add attendees"
  // link. Set when opening an event to manage it (list, home, take attendance).
  startAttendeesExpanded?: boolean
}

export const EventForm = ({
  mode,
  startExpanded,
  startAttendeesExpanded,
}: EventFormProps) => {
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
    isDirty,
  } = useEventForm({ mode, startExpanded, startAttendeesExpanded })

  // Show the attendee fields outright for everything except a brand-new public
  // event, where they sit behind an "Add attendees" link. The link is bypassed
  // once the user has opened it or already entered attendees — e.g. when they
  // filled in the quick-attendance form and then checked "Public event", we
  // don't want their entries to disappear.
  const showAttendees =
    showAttendeeSection || attendeesExpanded || attendeeCount > 0

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

  // The event detail fields. Rendered directly for a new event, or inside the
  // collapsible EventDetailsCard when editing a saved one.
  const detailFields = (
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
              placeholder={`Enter ${isConnection ? 'connection' : 'event'} name`}
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
                    field.state.value ? parseISO(field.state.value) : undefined
                  }
                  onValueChange={(date) => {
                    field.handleChange(date ? format(date, 'yyyy-MM-dd') : '')
                  }}
                  placeholder="Pick a date"
                  className={cn(field.state.meta.errors[0] && 'border-red-500')}
                />
                <FieldError
                  message={field.state.meta.errors[0]?.message}
                  className="mt-1"
                />
              </div>
              <Button type="button" variant="outline" onClick={setDateToToday}>
                Today
              </Button>
            </div>
          </div>
        )}
      </form.Field>

      {/* Public event toggle. Switching it on reveals the scheduled-event
          fields below. Hidden for coaching. */}
      {!isConnection && (
        <form.Field name="isPublic">
          {(field) => (
            <div className="flex items-center gap-2">
              <Switch
                id="isPublic"
                checked={field.state.value}
                onCheckedChange={(checked) => field.handleChange(checked)}
              />
              <Label htmlFor="isPublic" className="cursor-pointer">
                Public event
              </Label>
              <Popover>
                <PopoverTrigger asChild>
                  <button
                    type="button"
                    className="text-muted-foreground hover:text-foreground"
                    aria-label="About public events"
                  >
                    <Info className="h-4 w-4" />
                  </button>
                </PopoverTrigger>
                <PopoverContent className="w-auto max-w-xs text-sm">
                  Public events will be listed publicly on the website (coming
                  soon).
                </PopoverContent>
              </Popover>
            </div>
          )}
        </form.Field>
      )}

      {/* Scheduled-event fields: time, timezone, location, description.
          Revealed by the "Public event" switch and grouped in a nested
          sub-box. Kept mounted (via Collapse) so it slides shut as well as open;
          `active` defers the Google Maps load until it's actually shown. */}
      <Collapse
        open={showUpcomingFields}
        className={cn(!showUpcomingFields && '-mt-4')}
      >
        <ScheduledEventFields
          form={form}
          googlePlacesApiKey={googlePlacesApiKey}
          active={showUpcomingFields}
        />
      </Collapse>

      {/* Suppress Survey Checkbox */}
      <Collapse
        open={shouldShowSuppressSurveyCheckbox}
        className={cn(!shouldShowSuppressSurveyCheckbox && '-mt-4')}
      >
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
      </Collapse>
    </>
  )

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
      {/* Detail fields. Rendered directly for a new event; for a saved event
          they collapse into a card whose header doubles as the toggle, so
          attendees sit near the top. */}
      {eventId ? (
        <EventDetailsCard
          form={form}
          isConnection={isConnection}
          detailsExpanded={detailsExpanded}
          onToggle={() => setDetailsExpanded((v) => !v)}
        >
          {detailFields}
        </EventDetailsCard>
      ) : (
        <>
          {detailFields}
          {/* Divider between the detail fields and the attendee section so the
              two read as distinct sections. */}
          <hr className="border-border" />
        </>
      )}

      {/* Attendees/Coachees Section. For a brand-new public event attendance is
          usually recorded later, at the event, so the fields are hidden behind a
          small "Add attendees" link the user can open to record them up front. */}
      {showAttendees ? (
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
        <button
          type="button"
          onClick={() => setAttendeesExpanded(true)}
          className="self-start text-sm font-medium text-primary hover:underline"
        >
          Add {isConnection ? 'coachees' : 'attendees'}
        </button>
      )}

      {/* Save Button with Attendee/Coachee Count */}
      <div className="flex justify-between items-center">
        {showAttendees ? (
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
