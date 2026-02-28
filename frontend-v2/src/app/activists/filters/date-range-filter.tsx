'use client'

import { useState } from 'react'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { DatePicker } from '@/components/ui/date-picker'
import { Checkbox } from '@/components/ui/checkbox'
import { FilterChip } from './filter-chip'
import { useDraftFilter } from './filter-utils'
import { format } from 'date-fns'

interface DateRangeFilterProps {
  label: string
  /** URL-format range string: "2025-01-01..2025-06-01", "-180..", etc. */
  value?: string
  onChange: (value?: string) => void
  defaultOpen?: boolean
  removable?: boolean
  /** Label for the "or null" checkbox, e.g. "Include activists with no events". */
  nullLabel?: string
}

// --- Date helpers ---

function toUtcDate(dateString: string) {
  return new Date(`${dateString}T00:00:00`)
}

function toDateString(date?: Date) {
  return date
    ? `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
    : undefined
}

// --- Relative date helpers ---

/** Check if a value part is a relative day offset (integer, e.g. "-180"). */
function isRelative(v: string): boolean {
  return /^-?\d+$/.test(v)
}

type DateUnit = 'days' | 'weeks' | 'months'

/** Decompose an absolute day count into the largest clean unit. */
function decompose(absDays: number): { amount: number; unit: DateUnit } {
  if (absDays > 0 && absDays % 30 === 0) return { amount: absDays / 30, unit: 'months' }
  if (absDays > 0 && absDays % 7 === 0) return { amount: absDays / 7, unit: 'weeks' }
  return { amount: absDays, unit: 'days' }
}

/** Format an absolute day count as "6 months", "2 weeks", "45 days", or "today". */
function formatDays(absDays: number): string {
  if (absDays === 0) return 'today'
  const { amount, unit } = decompose(absDays)
  if (amount === 1) return `1 ${unit.slice(0, -1)}`
  return `${amount} ${unit}`
}

const UNIT_MULTIPLIER: Record<DateUnit, number> = { days: 1, weeks: 7, months: 30 }

// --- Parse / build URL range syntax ---

/**
 * Parses the URL date range syntax into component parts.
 *
 * Syntax: [gte]..[lt][|null]
 * Each bound can be an absolute date (YYYY-MM-DD) or a relative day offset
 * (negative integer, e.g. -180 for 180 days ago).
 *
 *   "2025-01-01..2025-06-01"       → absolute gte + lt
 *   "-180.."                        → relative gte (within last 180 days)
 *   "..-360|null"                   → relative lt (over 360 days ago) or null
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
  // "or null" only makes sense for open-ended ranges (one bound missing).
  const effectiveOrNull = orNull && !(gte && lt)
  if (!gte && !lt && !effectiveOrNull) return undefined
  if (!gte && !lt && effectiveOrNull) return 'null'
  const range = `${gte || ''}..${lt || ''}`
  return effectiveOrNull ? `${range}|null` : range
}

// --- Chip summary formatting ---

function formatDateRange(
  gte?: string,
  lt?: string,
  orNull?: boolean,
): string | undefined {
  const gteRel = gte && isRelative(gte)
  const ltRel = lt && isRelative(lt)

  let rangeText: string | undefined

  // Both relative — format each side independently so "today" isn't followed by "ago"
  if (gteRel && ltRel) {
    const gteAbs = Math.abs(parseInt(gte!, 10))
    const ltAbs = Math.abs(parseInt(lt!, 10))
    const gteStr = gteAbs === 0 ? 'today' : `${formatDays(gteAbs)} ago`
    const ltStr = ltAbs === 0 ? 'today' : `${formatDays(ltAbs)} ago`
    rangeText = `${gteStr} – ${ltStr}`
  } else if (gteRel && !lt) {
    const gteAbs = Math.abs(parseInt(gte!, 10))
    rangeText = gteAbs === 0 ? 'Today onward' : `Last ${formatDays(gteAbs)}`
  } else if (ltRel && !gte) {
    const ltAbs = Math.abs(parseInt(lt!, 10))
    rangeText = ltAbs === 0 ? 'Before today' : `Over ${formatDays(ltAbs)} ago`
  } else if (gteRel || ltRel) {
    // Mixed: format each side independently
    const parts: string[] = []
    if (gte) {
      parts.push(
        gteRel
          ? `≥ ${formatDays(Math.abs(parseInt(gte, 10)))} ago`
          : `≥ ${format(toUtcDate(gte), 'MMM d, yyyy')}`,
      )
    }
    if (lt) {
      parts.push(
        ltRel
          ? `< ${formatDays(Math.abs(parseInt(lt, 10)))} ago`
          : `< ${format(toUtcDate(lt), 'MMM d, yyyy')}`,
      )
    }
    rangeText = parts.join(', ')
  } else {
    // Both absolute (original logic)
    const currentYear = new Date().getFullYear()
    if (gte && lt) {
      const gteDate = toUtcDate(gte)
      const ltDate = toUtcDate(lt)
      const bothCurrentYear =
        gteDate.getFullYear() === currentYear &&
        ltDate.getFullYear() === currentYear
      rangeText = bothCurrentYear
        ? `${format(gteDate, 'MMM d')} – ${format(ltDate, 'MMM d')}`
        : `${format(gteDate, 'MMM d, yyyy')} – ${format(ltDate, 'MMM d, yyyy')}`
    } else if (gte) {
      rangeText = `On or after ${format(toUtcDate(gte), 'MMM d, yyyy')}`
    } else if (lt) {
      rangeText = `Before ${format(toUtcDate(lt), 'MMM d, yyyy')}`
    }
  }

  if (orNull) {
    return rangeText ? `${rangeText} or none` : 'None'
  }
  return rangeText
}

// --- Relative input helpers ---

function parseRelativeInput(
  v?: string,
): { amount: number; unit: DateUnit } | undefined {
  if (!v || !isRelative(v)) return undefined
  const absDays = Math.abs(parseInt(v, 10))
  return decompose(absDays)
}

function buildRelativeValue(amount: number, unit: DateUnit): string {
  const days = amount * UNIT_MULTIPLIER[unit]
  return String(-days)
}

// --- Component ---

type Mode = 'absolute' | 'relative'

function detectMode(gte?: string, lt?: string): Mode | undefined {
  if (gte && isRelative(gte)) return 'relative'
  if (lt && isRelative(lt)) return 'relative'
  if (gte || lt) return 'absolute'
  return undefined
}

export function DateRangeFilter({
  label,
  value,
  onChange,
  defaultOpen,
  removable,
  nullLabel,
}: DateRangeFilterProps) {
  const [draft, setDraft, onOpenChange] = useDraftFilter(value, onChange)

  const { gte, lt, orNull } = parseValue(draft)
  const hasDraft = !!draft
  const detectedMode = detectMode(gte, lt)
  const [preferredMode, setPreferredMode] = useState<Mode>('relative')
  const mode = detectedMode ?? preferredMode

  const handleModeSwitch = (newMode: Mode) => {
    setPreferredMode(newMode)
    if (detectedMode && newMode !== detectedMode) {
      // Clear values when switching modes (preserve orNull)
      setDraft(orNull ? 'null' : undefined)
    }
  }

  // Absolute handlers
  const handleAbsGte = (date?: Date) =>
    setDraft(buildValue(toDateString(date), lt, orNull))
  const handleAbsLt = (date?: Date) =>
    setDraft(buildValue(gte, toDateString(date), orNull))

  // Relative handlers
  const gteRel = parseRelativeInput(gte)
  const ltRel = parseRelativeInput(lt)

  const handleRelGte = (amount: number | undefined, unit: DateUnit) =>
    setDraft(
      buildValue(
        amount != null ? buildRelativeValue(amount, unit) : undefined,
        lt,
        orNull,
      ),
    )
  const handleRelLt = (amount: number | undefined, unit: DateUnit) =>
    setDraft(
      buildValue(
        gte,
        amount != null ? buildRelativeValue(amount, unit) : undefined,
        orNull,
      ),
    )

  // Summary is derived from the committed value, not the draft.
  const committed = parseValue(value)
  const summary = value
    ? formatDateRange(committed.gte, committed.lt, committed.orNull)
    : undefined

  return (
    <FilterChip
      label={label}
      summary={summary}
      onClear={() => onChange(undefined)}
      defaultOpen={defaultOpen}
      removable={removable}
      popoverClassName="w-80"
      onOpenChange={onOpenChange}
    >
      <div className="space-y-4">
        {/* Mode toggle */}
        <div className="flex rounded-md border overflow-hidden text-sm">
          <button
            className={`flex-1 px-3 py-1.5 transition-colors ${
              mode === 'absolute'
                ? 'bg-primary text-primary-foreground'
                : 'hover:bg-muted'
            }`}
            onClick={() => handleModeSwitch('absolute')}
          >
            Absolute
          </button>
          <button
            className={`flex-1 px-3 py-1.5 transition-colors border-l ${
              mode === 'relative'
                ? 'bg-primary text-primary-foreground'
                : 'hover:bg-muted'
            }`}
            onClick={() => handleModeSwitch('relative')}
          >
            Relative
          </button>
        </div>

        {mode === 'absolute' ? (
          <>
            <div className="space-y-2">
              <Label className="text-sm font-medium">On or after</Label>
              <DatePicker
                value={gte && !isRelative(gte) ? toUtcDate(gte) : undefined}
                onValueChange={handleAbsGte}
                placeholder="Select start date"
              />
            </div>
            <div className="space-y-2">
              <Label className="text-sm font-medium">Before</Label>
              <DatePicker
                value={lt && !isRelative(lt) ? toUtcDate(lt) : undefined}
                onValueChange={handleAbsLt}
                placeholder="Select end date"
              />
            </div>
          </>
        ) : (
          <>
            <RelativeInput
              label="On or after"
              amount={gteRel?.amount}
              unit={gteRel?.unit ?? 'months'}
              onChange={handleRelGte}
            />
            <RelativeInput
              label="Before"
              amount={ltRel?.amount}
              unit={ltRel?.unit ?? 'months'}
              onChange={handleRelLt}
            />
          </>
        )}

        {nullLabel && (() => {
          const bothBoundsSet = !!gte && !!lt
          return (
            <div className="flex items-center gap-2">
              <Checkbox
                id="or-null"
                checked={orNull && !bothBoundsSet}
                disabled={bothBoundsSet}
                onCheckedChange={(checked) =>
                  setDraft(buildValue(gte, lt, !!checked))
                }
              />
              <label
                htmlFor="or-null"
                className={`text-sm ${bothBoundsSet ? 'text-muted-foreground' : ''}`}
              >
                {nullLabel}
              </label>
            </div>
          )
        })()}

        {hasDraft && (
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

function RelativeInput({
  label: inputLabel,
  amount,
  unit,
  onChange,
}: {
  label: string
  amount?: number
  unit: DateUnit
  onChange: (amount: number | undefined, unit: DateUnit) => void
}) {
  const [localUnit, setLocalUnit] = useState<DateUnit>(unit)

  const handleAmountChange = (newAmount: string) => {
    if (newAmount === '') {
      onChange(undefined, localUnit)
      return
    }
    const n = parseInt(newAmount, 10)
    onChange(isNaN(n) || n < 0 ? undefined : n, localUnit)
  }

  const handleUnitChange = (newUnit: DateUnit) => {
    setLocalUnit(newUnit)
    if (amount != null) {
      onChange(amount, newUnit)
    }
  }

  return (
    <div className="space-y-2">
      <Label className="text-sm font-medium">{inputLabel}</Label>
      <div className="flex items-center gap-2">
        <Input
          type="number"
          min={0}
          className="w-20"
          value={amount ?? ''}
          onChange={(e) => handleAmountChange(e.target.value)}
          placeholder="0"
        />
        <Select value={localUnit} onValueChange={(v) => handleUnitChange(v as DateUnit)}>
          <SelectTrigger className="w-28">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="days">days ago</SelectItem>
            <SelectItem value="weeks">weeks ago</SelectItem>
            <SelectItem value="months">months ago</SelectItem>
          </SelectContent>
        </Select>
      </div>
    </div>
  )
}
