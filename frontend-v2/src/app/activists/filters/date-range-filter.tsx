'use client'

import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { DatePicker } from '@/components/ui/date-picker'
import { FilterChip } from './filter-chip'
import { format } from 'date-fns'

interface DateRangeFilterProps {
  label: string
  /** URL-format range string: "2025-01-01..2025-06-01", "..2025-06-01|null", etc. */
  value?: string
  onChange: (value?: string) => void
  defaultOpen?: boolean
  removable?: boolean
}

function toUtcDate(dateString: string) {
  return new Date(`${dateString}T00:00:00`)
}

function toDateString(date?: Date) {
  return date
    ? `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
    : undefined
}

/**
 * Parses the URL date range syntax into component parts.
 *
 * Syntax: [gte]..[lt][|null]
 *   "2025-01-01..2025-06-01"       → gte + lt (between two dates)
 *   "2025-01-01.."                  → gte only (on or after)
 *   "..2025-06-01"                  → lt only (before)
 *   "..2025-06-01|null"             → lt, including NULL values
 *   "2025-01-01..|null"             → gte, including NULL values
 *   "2025-01-01..2025-06-01|null"   → gte + lt, including NULL values
 *   "null"                          → only NULL values
 */
function parseValue(value?: string): {
  gte?: string
  lt?: string
  orNull: boolean
} {
  if (!value) return { orNull: false }
  let orNull = false
  let range = value
  if (range.endsWith('|null')) {
    orNull = true
    range = range.slice(0, -5)
  }
  if (range === 'null') return { orNull: true }
  const parts = range.split('..')
  if (parts.length !== 2) return { orNull }
  return {
    gte: parts[0] || undefined,
    lt: parts[1] || undefined,
    orNull,
  }
}

/** Inverse of parseValue — see parseValue for full syntax. */
function buildValue(
  gte?: string,
  lt?: string,
  orNull?: boolean,
): string | undefined {
  if (!gte && !lt && !orNull) return undefined
  if (!gte && !lt && orNull) return 'null'
  const range = `${gte || ''}..${lt || ''}`
  return orNull ? `${range}|null` : range
}

function formatDateRange(
  gte?: string,
  lt?: string,
  orNull?: boolean,
): string | undefined {
  const currentYear = new Date().getFullYear()
  let rangeText: string | undefined
  if (gte && lt) {
    const gteDate = toUtcDate(gte)
    const ltDate = toUtcDate(lt)
    const bothCurrentYear =
      gteDate.getFullYear() === currentYear &&
      ltDate.getFullYear() === currentYear
    rangeText = bothCurrentYear
      ? `${format(gteDate, 'MMM d')} - ${format(ltDate, 'MMM d')}`
      : `${format(gteDate, 'MMM d, yyyy')} - ${format(ltDate, 'MMM d, yyyy')}`
  } else if (gte) {
    rangeText = `On or after ${format(toUtcDate(gte), 'MMM d, yyyy')}`
  } else if (lt) {
    rangeText = `Before ${format(toUtcDate(lt), 'MMM d, yyyy')}`
  }
  if (orNull) {
    return rangeText ? `${rangeText} or none` : 'None'
  }
  return rangeText
}

export function DateRangeFilter({
  label,
  value,
  onChange,
  defaultOpen,
  removable,
}: DateRangeFilterProps) {
  const { gte, lt, orNull } = parseValue(value)
  const hasFilter = !!value

  return (
    <FilterChip
      label={label}
      summary={hasFilter ? formatDateRange(gte, lt, orNull) : undefined}
      onClear={() => onChange(undefined)}
      defaultOpen={defaultOpen}
      removable={removable}
      popoverClassName="w-80"
    >
      <div className="space-y-4">
        <div className="space-y-2">
          <Label className="text-sm font-medium">On or after</Label>
          <DatePicker
            value={gte ? toUtcDate(gte) : undefined}
            onValueChange={(date) =>
              onChange(buildValue(toDateString(date), lt, orNull))
            }
            placeholder="Select start date"
          />
        </div>
        <div className="space-y-2">
          <Label className="text-sm font-medium">Before</Label>
          <DatePicker
            value={lt ? toUtcDate(lt) : undefined}
            onValueChange={(date) =>
              onChange(buildValue(gte, toDateString(date), orNull))
            }
            placeholder="Select end date"
          />
        </div>
        {hasFilter && (
          <Button
            variant="outline"
            size="sm"
            className="w-full"
            onClick={() => onChange(undefined)}
          >
            Clear dates
          </Button>
        )}
      </div>
    </FilterChip>
  )
}
