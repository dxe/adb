'use client'

import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { FilterChip } from './filter-chip'
import { DateRangeFields, formatDateRange } from './date-range-filter'
import { useDraftFilter } from '@/hooks/use-draft-filter'
import {
  FIRST_EVENT_TYPES,
  type DateRangeFilterValue,
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

/**
 * "Include activists with no events" applies when the date range has exactly
 * one bound, or has no bounds but event types are selected (then it means
 * "first event was one of these types, or no events").
 */
function isNullOptionEnabled(
  date?: DateRangeFilterValue,
  types?: EventTypeFilterValue,
): boolean {
  const boundCount = (date?.gte ? 1 : 0) + (date?.lt ? 1 : 0)
  if (boundCount === 1) return true
  return boundCount === 0 && !!types && types.values.length > 0
}

function normalizeFirstEvent(
  value: FirstEventFilterValue,
): FirstEventFilterValue | undefined {
  const types =
    value.types && value.types.values.length > 0 ? value.types : undefined
  let date = value.date
  if (date?.orNull && !isNullOptionEnabled(date, types)) {
    date = { gte: date.gte, lt: date.lt }
  }
  if (date && !date.gte && !date.lt && !date.orNull) date = undefined
  if (!date && !types) return undefined
  return { date, types }
}

export function formatFirstEvent(
  value?: FirstEventFilterValue,
): string | undefined {
  if (!value) return undefined
  const types = value.types
  const typesText =
    types && types.values.length > 0
      ? `${types.mode === 'exclude' ? 'not ' : ''}${types.values.join(', ')}`
      : undefined
  const dateText = formatDateRange(
    value.date && { gte: value.date.gte, lt: value.date.lt },
  )
  const text = [typesText, dateText].filter(Boolean).join(' · ')
  if (value.date?.orNull) return text ? `${text} or none` : 'None'
  return text || undefined
}

export function FirstEventFilter({
  value,
  onChange,
  defaultOpen,
  removable,
}: FirstEventFilterProps) {
  const [draft, setDraft, onOpenChange] = useDraftFilter(value, onChange)
  const date = draft?.date
  const types = draft?.types
  const mode = types?.mode ?? 'include'
  const selected = new Set<FirstEventTypeValue>(types?.values ?? [])

  const update = (next: FirstEventFilterValue) =>
    setDraft(normalizeFirstEvent(next))

  const setTypes = (
    nextMode: EventTypeFilterValue['mode'],
    values: Set<FirstEventTypeValue>,
  ) =>
    update({
      date,
      // Keep the list in display order rather than click order.
      types: {
        mode: nextMode,
        values: FIRST_EVENT_TYPES.filter((t) => values.has(t)),
      },
    })

  const handleToggle = (type: FirstEventTypeValue) => {
    const next = new Set(selected)
    if (next.has(type)) {
      next.delete(type)
    } else {
      next.add(type)
    }
    setTypes(mode, next)
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
          onChange={(d) => update({ date: d, types })}
          nullLabel="Include activists with no events"
          nullOptionEnabled={isNullOptionEnabled(date, types)}
          nullOptionDisabledHint="Set exactly one date bound, or select event types, to enable this option."
        />

        <div className="space-y-2 border-t pt-4">
          <div className="flex items-center justify-between">
            <h4 className="text-sm font-medium">Event type</h4>
            <Select
              value={mode}
              onValueChange={(v) =>
                setTypes(v as EventTypeFilterValue['mode'], selected)
              }
            >
              <SelectTrigger
                className="h-7 w-[100px] text-xs"
                aria-label="Event type mode"
              >
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="include">Include</SelectItem>
                <SelectItem value="exclude">Exclude</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="grid grid-cols-2 gap-x-2">
            {FIRST_EVENT_TYPES.map((type) => {
              const isSelected = selected.has(type)
              return (
                <button
                  type="button"
                  key={type}
                  aria-pressed={isSelected}
                  className="flex w-full items-center gap-2 rounded px-2 py-1 text-left text-sm hover:bg-muted transition-colors"
                  onClick={() => handleToggle(type)}
                >
                  <span
                    className={`flex h-4 w-4 shrink-0 items-center justify-center rounded border text-xs font-bold ${
                      isSelected
                        ? mode === 'include'
                          ? 'bg-primary text-primary-foreground border-primary'
                          : 'bg-destructive text-destructive-foreground border-destructive'
                        : 'border-input'
                    }`}
                  >
                    {isSelected ? (mode === 'include' ? '+' : '-') : ''}
                  </span>
                  {type}
                </button>
              )
            })}
          </div>
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
