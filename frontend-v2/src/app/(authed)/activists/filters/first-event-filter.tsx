'use client'

import { useId } from 'react'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { FilterChip } from './filter-chip'
import {
  DateRangeFields,
  NullOptionCheckbox,
  formatDateRange,
} from './date-range-filter'
import { useDraftFilter } from '@/hooks/use-draft-filter'
import {
  FIRST_EVENT_TYPES,
  type DateRangeValue,
  type EventTypeFilterValue,
  type FirstEventFilterValue,
  type FirstEventTypeValue,
} from '../filter-types'

interface FirstEventFilterProps {
  value?: FirstEventFilterValue
  onChange: (value?: FirstEventFilterValue) => void
  defaultOpen?: boolean
  removable?: boolean
}

// Select value for "no type condition". Radix Select items can't use ''.
const ANY_TYPE = 'any'
type TypeModeOption = EventTypeFilterValue['mode'] | typeof ANY_TYPE

const hasTypes = (types?: EventTypeFilterValue) =>
  !!types && types.values.length > 0

/**
 * "Include activists with no events" applies when the date range has exactly
 * one bound, or has no bounds but event types are selected (then it means
 * "first event was one of these types, or no events").
 */
function isNullOptionEnabled(
  date?: DateRangeValue,
  types?: EventTypeFilterValue,
): boolean {
  const boundCount = (date?.gte ? 1 : 0) + (date?.lt ? 1 : 0)
  if (boundCount === 1) return true
  return boundCount === 0 && hasTypes(types)
}

/**
 * Drops parts of a draft that don't filter anything: a type mode with no
 * types picked, or "no events" when the checkbox is disabled.
 */
function normalizeFirstEvent(
  value?: FirstEventFilterValue,
): FirstEventFilterValue | undefined {
  if (!value) return undefined
  const date = value.date?.gte || value.date?.lt ? value.date : undefined
  const types = hasTypes(value.types) ? value.types : undefined
  const orNull = (value.orNull && isNullOptionEnabled(date, types)) || undefined
  if (!date && !types && !orNull) return undefined
  return { date, types, orNull }
}

export function formatFirstEvent(
  value?: FirstEventFilterValue,
): string | undefined {
  if (!value) return undefined
  const types = value.types
  const typesText = hasTypes(types)
    ? `${types!.mode === 'exclude' ? 'not ' : ''}${types!.values.join(', ')}`
    : undefined
  const dateText = formatDateRange(value.date)
  const text = [typesText, dateText].filter(Boolean).join(' · ')
  if (value.orNull) return text ? `${text} or none` : 'None'
  return text || undefined
}

export function FirstEventFilter({
  value,
  onChange,
  defaultOpen,
  removable,
}: FirstEventFilterProps) {
  // The draft may hold a type mode with no types picked yet, so normalize only
  // when committing. Skip committing an empty draft over an empty value, which
  // would remove a just-added chip.
  const commit = (next?: FirstEventFilterValue) => {
    const normalized = normalizeFirstEvent(next)
    if (normalized || value) onChange(normalized)
  }
  const [draft, setDraft, onOpenChange] = useDraftFilter(value, commit)
  const date = draft?.date
  const types = draft?.types
  const orNull = !!draft?.orNull
  const selected = new Set<FirstEventTypeValue>(types?.values ?? [])
  const typeModeId = useId()

  const update = (next: Partial<FirstEventFilterValue>) =>
    setDraft({ date, types, orNull: orNull || undefined, ...next })

  const handleTypeModeChange = (mode: TypeModeOption) =>
    update({
      types:
        mode === ANY_TYPE ? undefined : { mode, values: types?.values ?? [] },
    })

  const handleToggle = (type: FirstEventTypeValue, checked: boolean) => {
    if (!types) return
    const next = new Set(selected)
    if (checked) {
      next.add(type)
    } else {
      next.delete(type)
    }
    // Keep the list in display order rather than click order.
    update({
      types: {
        mode: types.mode,
        values: FIRST_EVENT_TYPES.filter((t) => next.has(t)),
      },
    })
  }

  return (
    <FilterChip
      label="First event"
      summary={formatFirstEvent(value)}
      onClear={() => onChange(undefined)}
      defaultOpen={defaultOpen}
      removable={removable}
      popoverClassName="w-96"
      onOpenChange={onOpenChange}
    >
      <div className="space-y-4">
        <DateRangeFields
          value={date}
          onChange={(nextDate) => update({ date: nextDate })}
        />

        <div className="space-y-3 border-t pt-4">
          <div className="flex items-center gap-3">
            <label htmlFor={typeModeId} className="text-sm font-medium">
              First event type is
            </label>
            <Select
              value={types?.mode ?? ANY_TYPE}
              onValueChange={(v) => handleTypeModeChange(v as TypeModeOption)}
            >
              <SelectTrigger id={typeModeId} className="h-8 w-32">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={ANY_TYPE} aria-label="Any type">
                  {/* Blank, so the sentence reads as unfinished until a mode
                      is picked. */}
                  {' '}
                </SelectItem>
                <SelectItem value="include">one of</SelectItem>
                <SelectItem value="exclude">not one of</SelectItem>
              </SelectContent>
            </Select>
          </div>
          {types && (
            <div className="grid grid-cols-2 gap-x-3 gap-y-2">
              {FIRST_EVENT_TYPES.map((type) => (
                <EventTypeCheckbox
                  key={type}
                  type={type}
                  checked={selected.has(type)}
                  onCheckedChange={(checked) => handleToggle(type, checked)}
                />
              ))}
            </div>
          )}
        </div>

        <div className="border-t pt-4">
          <NullOptionCheckbox
            label="Include activists with no events"
            checked={orNull}
            enabled={isNullOptionEnabled(date, types)}
            disabledHint="Set exactly one date bound, or pick event types, to enable this option."
            onCheckedChange={(checked) => update({ orNull: checked })}
          />
        </div>

        {draft && (
          <Button
            variant="outline"
            size="sm"
            className="w-full"
            onClick={() => setDraft(undefined)}
          >
            Clear
          </Button>
        )}
      </div>
    </FilterChip>
  )
}

function EventTypeCheckbox({
  type,
  checked,
  onCheckedChange,
}: {
  type: FirstEventTypeValue
  checked: boolean
  onCheckedChange: (checked: boolean) => void
}) {
  const id = useId()
  return (
    <div className="flex items-start gap-2">
      <Checkbox
        id={id}
        className="mt-0.5"
        checked={checked}
        onCheckedChange={(c) => onCheckedChange(c === true)}
      />
      <label htmlFor={id} className="text-sm leading-tight">
        {type}
      </label>
    </div>
  )
}
