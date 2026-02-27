'use client'

import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { DatePicker } from '@/components/ui/date-picker'
import { ChevronDown, X } from 'lucide-react'
import { format } from 'date-fns'

interface DateRangeFilterProps {
  label: string
  /** URL-format range string: "2025-01-01..2025-06-01", "..2025-06-01|null", etc. */
  value?: string
  onChange: (value?: string) => void
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
function buildValue(gte?: string, lt?: string, orNull?: boolean): string | undefined {
  if (!gte && !lt && !orNull) return undefined
  if (!gte && !lt && orNull) return 'null'
  const range = `${gte || ''}..${lt || ''}`
  return orNull ? `${range}|null` : range
}

export function DateRangeFilter({ label, value, onChange }: DateRangeFilterProps) {
  const { gte, lt } = parseValue(value)
  const hasFilter = !!value

  const formatDateRange = () => {
    const currentYear = new Date().getFullYear()
    if (gte && lt) {
      const gteDate = toUtcDate(gte)
      const ltDate = toUtcDate(lt)
      const bothCurrentYear =
        gteDate.getFullYear() === currentYear &&
        ltDate.getFullYear() === currentYear
      return bothCurrentYear
        ? `${format(gteDate, 'MMM d')} - ${format(ltDate, 'MMM d')}`
        : `${format(gteDate, 'MMM d, yyyy')} - ${format(ltDate, 'MMM d, yyyy')}`
    } else if (gte) {
      return `On or after ${format(toUtcDate(gte), 'MMM d, yyyy')}`
    } else if (lt) {
      return `Before ${format(toUtcDate(lt), 'MMM d, yyyy')}`
    }
    return null
  }

  const handleClear = () => onChange(undefined)

  const handleGteChange = (date?: Date) => {
    onChange(buildValue(toDateString(date), lt))
  }

  const handleLtChange = (date?: Date) => {
    onChange(buildValue(gte, toDateString(date)))
  }

  return (
    <div className="flex shrink-0 items-stretch rounded-md border bg-card overflow-hidden h-12">
      <Popover>
        <PopoverTrigger asChild>
          <button className="flex flex-col items-start justify-center px-3 hover:bg-muted transition-colors h-full">
            {!hasFilter ? (
              <div className="flex items-center gap-1">
                <span className="text-sm">{label}</span>
                <ChevronDown className="h-3 w-3 text-muted-foreground" />
              </div>
            ) : (
              <>
                <span className="text-xs text-muted-foreground">{label}</span>
                <div className="flex items-center gap-1">
                  <span className="text-sm">{formatDateRange()}</span>
                  <ChevronDown className="h-3 w-3 text-muted-foreground" />
                </div>
              </>
            )}
          </button>
        </PopoverTrigger>
        <PopoverContent className="w-80">
          <div className="space-y-4">
            <div className="space-y-2">
              <Label className="text-sm font-medium">On or after</Label>
              <DatePicker
                value={gte ? toUtcDate(gte) : undefined}
                onValueChange={handleGteChange}
                placeholder="Select start date"
              />
            </div>
            <div className="space-y-2">
              <Label className="text-sm font-medium">Before</Label>
              <DatePicker
                value={lt ? toUtcDate(lt) : undefined}
                onValueChange={handleLtChange}
                placeholder="Select end date"
              />
            </div>
            {hasFilter && (
              <Button
                variant="outline"
                size="sm"
                className="w-full"
                onClick={handleClear}
              >
                Clear dates
              </Button>
            )}
          </div>
        </PopoverContent>
      </Popover>
      {hasFilter && (
        <button
          onClick={handleClear}
          className="border-l px-2 hover:bg-muted transition-colors"
          aria-label="Clear filter"
        >
          <X className="h-4 w-4" />
        </button>
      )}
    </div>
  )
}
