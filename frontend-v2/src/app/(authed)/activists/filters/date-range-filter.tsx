'use client'

import { useEffect, useId, useState } from 'react'
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
import { CircleHelp } from 'lucide-react'
import { FilterChip } from './filter-chip'
import { useDraftFilter } from '@/hooks/use-draft-filter'
import type {
  AbsoluteDateBound,
  DateRangeBoundValue,
  DateRangeFilterValue,
  RelativeDateBound,
} from '../filter-types'
import { parseSafeInteger } from '../number-utils'
import {
  datePickerValueToYmd,
  formatYmdForActivists,
  getCurrentYearInActivistsTimeZone,
  ymdToDatePickerValue,
} from '../date-time'

interface DateRangeFilterProps {
  label: string
  value?: DateRangeFilterValue
  onChange: (value?: DateRangeFilterValue) => void
  defaultOpen?: boolean
  removable?: boolean
  /** Label for the "or null" checkbox, e.g. "Include activists with no events". */
  nullLabel?: string
}

type DateUnit = 'days' | 'weeks' | 'months'
type Mode = 'absolute' | 'relative'

const UNIT_MULTIPLIER: Record<DateUnit, number> = {
  days: 1,
  weeks: 7,
  months: 30,
}

function isRelativeBound(
  value?: DateRangeBoundValue,
): value is RelativeDateBound {
  return value?.mode === 'relative'
}

function isAbsoluteBound(
  value?: DateRangeBoundValue,
): value is AbsoluteDateBound {
  return value?.mode === 'absolute'
}

function getBoundsMode(
  gte?: DateRangeBoundValue,
  lt?: DateRangeBoundValue,
): Mode | undefined {
  return gte?.mode ?? lt?.mode
}

/** Decompose an absolute day count into the largest clean unit. */
function decomposeDayCount(absDays: number): {
  amount: number
  unit: DateUnit
} {
  if (absDays > 0 && absDays % 30 === 0)
    return { amount: absDays / 30, unit: 'months' }
  if (absDays > 0 && absDays % 7 === 0)
    return { amount: absDays / 7, unit: 'weeks' }
  return { amount: absDays, unit: 'days' }
}

/** Format an absolute day count as "6 months", "2 weeks", "45 days", or "today". */
function formatDayCount(absDays: number): string {
  if (absDays === 0) return 'today'
  const { amount, unit } = decomposeDayCount(absDays)
  if (amount === 1) return `1 ${unit.slice(0, -1)}`
  return `${amount} ${unit}`
}

function normalizeDateRange(
  value?: DateRangeFilterValue,
): DateRangeFilterValue | undefined {
  if (!value) return undefined

  const gte = value.gte
  const lt = value.lt
  const orNull = !!value.orNull && !(gte && lt)

  if (!gte && !lt && !orNull) return undefined

  return {
    gte,
    lt,
    orNull: orNull || undefined,
  }
}

function formatRelativeReferenceLabel(daysOffset: number): string {
  if (daysOffset === 0) return 'today'
  const distance = formatDayCount(Math.abs(daysOffset))
  return daysOffset < 0 ? `${distance} ago` : `${distance} from now`
}

function formatAbsoluteRange(
  gte?: AbsoluteDateBound,
  lt?: AbsoluteDateBound,
): string | undefined {
  const currentYear = getCurrentYearInActivistsTimeZone()
  if (gte && lt) {
    if (gte.date >= lt.date) return 'No matching dates'
    const gteYear = Number(gte.date.slice(0, 4))
    const ltYear = Number(lt.date.slice(0, 4))
    const bothCurrentYear = gteYear === currentYear && ltYear === currentYear
    return bothCurrentYear
      ? `${formatYmdForActivists(gte.date, { includeYear: false })} – ${formatYmdForActivists(lt.date, { includeYear: false })}`
      : `${formatYmdForActivists(gte.date)} – ${formatYmdForActivists(lt.date)}`
  }
  if (gte) return `On or after ${formatYmdForActivists(gte.date)}`
  if (lt) return `Before ${formatYmdForActivists(lt.date)}`
  return undefined
}

export function formatDateRange(
  value?: DateRangeFilterValue,
): string | undefined {
  const gte = value?.gte
  const lt = value?.lt
  const orNull = value?.orNull
  const boundsMode = getBoundsMode(gte, lt)

  let rangeText: string | undefined

  if (boundsMode === 'relative') {
    const relGte = isRelativeBound(gte) ? gte : undefined
    const relLt = isRelativeBound(lt) ? lt : undefined

    if (relGte && relLt) {
      if (relGte.daysOffset >= relLt.daysOffset) {
        rangeText = 'No matching dates'
      } else {
        const gteStr = formatRelativeReferenceLabel(relGte.daysOffset)
        const ltStr = formatRelativeReferenceLabel(relLt.daysOffset)
        rangeText = `${gteStr} – ${ltStr}`
      }
    } else if (relGte) {
      if (relGte.daysOffset === 0) {
        rangeText = 'Today onward'
      } else {
        rangeText = `On or after ${formatRelativeReferenceLabel(relGte.daysOffset)}`
      }
    } else if (relLt) {
      if (relLt.daysOffset === 0) {
        rangeText = 'Before today'
      } else if (relLt.daysOffset < 0) {
        rangeText = `Over ${formatDayCount(Math.abs(relLt.daysOffset))} ago`
      } else {
        rangeText = `Before ${formatRelativeReferenceLabel(relLt.daysOffset)}`
      }
    }
  } else {
    rangeText = formatAbsoluteRange(
      isAbsoluteBound(gte) ? gte : undefined,
      isAbsoluteBound(lt) ? lt : undefined,
    )
  }

  if (orNull) {
    return rangeText ? `${rangeText} or none` : 'None'
  }
  return rangeText
}

function parseRelativeInput(
  value?: DateRangeBoundValue,
): { amount: number; unit: DateUnit } | undefined {
  if (!isRelativeBound(value)) return undefined
  return decomposeDayCount(Math.abs(value.daysOffset))
}

function toAbsoluteBound(date?: Date): AbsoluteDateBound | undefined {
  const ymd = date ? datePickerValueToYmd(date) : undefined
  return ymd ? { mode: 'absolute', date: ymd } : undefined
}

function buildRelativeBound(amount: number, unit: DateUnit): RelativeDateBound {
  return {
    mode: 'relative',
    daysOffset: -(amount * UNIT_MULTIPLIER[unit]),
  }
}

function inferModeFromBounds(
  gte?: DateRangeBoundValue,
  lt?: DateRangeBoundValue,
): Mode | undefined {
  return getBoundsMode(gte, lt)
}

export function DateRangeFilter({
  label,
  value,
  onChange,
  defaultOpen,
  removable,
  nullLabel,
}: DateRangeFilterProps) {
  const [draft, setDraft, onDraftOpenChange] = useDraftFilter(value, onChange)

  const gte = draft?.gte
  const lt = draft?.lt
  const orNull = !!draft?.orNull
  const hasGte = !!gte
  const hasLt = !!lt
  const exactlyOneBoundSet = hasGte !== hasLt
  const nullOptionDisabled = !exactlyOneBoundSet
  const orNullId = useId()
  const hasDraft = !!draft
  const detectedMode = inferModeFromBounds(gte, lt)
  const [preferredMode, setPreferredMode] = useState<Mode>('relative')
  const mode = detectedMode ?? preferredMode

  const onOpenChange = (open: boolean) => {
    if (open) {
      setPreferredMode(inferModeFromBounds(value?.gte, value?.lt) ?? 'relative')
    }
    onDraftOpenChange(open)
  }

  const updateDraft = (next: DateRangeFilterValue | undefined) => {
    setDraft(normalizeDateRange(next))
  }

  const handleModeSwitch = (newMode: Mode) => {
    setPreferredMode(newMode)
    if (detectedMode && newMode !== detectedMode) {
      // Clear values when switching modes (preserve orNull).
      updateDraft(orNull ? { orNull: true } : undefined)
    }
  }

  const handleAbsoluteStartChange = (date?: Date) =>
    updateDraft({
      gte: toAbsoluteBound(date),
      lt,
      orNull,
    })
  const handleAbsoluteEndChange = (date?: Date) =>
    updateDraft({
      gte,
      lt: toAbsoluteBound(date),
      orNull,
    })

  const gteRel = parseRelativeInput(gte)
  const ltRel = parseRelativeInput(lt)

  const handleRelGte = (amount: number | undefined, unit: DateUnit) =>
    updateDraft({
      gte: amount != null ? buildRelativeBound(amount, unit) : undefined,
      lt,
      orNull,
    })
  const handleRelLt = (amount: number | undefined, unit: DateUnit) =>
    updateDraft({
      gte,
      lt: amount != null ? buildRelativeBound(amount, unit) : undefined,
      orNull,
    })

  const summary = formatDateRange(value)

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
        <div className="flex rounded-md border overflow-hidden text-sm">
          <button
            type="button"
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
            type="button"
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
                value={
                  isAbsoluteBound(gte)
                    ? ymdToDatePickerValue(gte.date)
                    : undefined
                }
                onValueChange={handleAbsoluteStartChange}
                placeholder="Select start date"
              />
            </div>
            <div className="space-y-2">
              <Label className="text-sm font-medium">Before</Label>
              <DatePicker
                value={
                  isAbsoluteBound(lt)
                    ? ymdToDatePickerValue(lt.date)
                    : undefined
                }
                onValueChange={handleAbsoluteEndChange}
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

        {nullLabel && (
          <div className="flex items-center gap-2">
            <Checkbox
              id={orNullId}
              checked={orNull && exactlyOneBoundSet}
              disabled={nullOptionDisabled}
              onCheckedChange={(checked) =>
                updateDraft({ gte, lt, orNull: checked === true })
              }
            />
            <label
              htmlFor={orNullId}
              className={`text-sm ${nullOptionDisabled ? 'text-muted-foreground' : ''}`}
            >
              {nullLabel}
            </label>
            {nullOptionDisabled && (
              <span title="Set exactly one bound to enable this option.">
                <CircleHelp className="h-3.5 w-3.5 text-muted-foreground shrink-0" />
              </span>
            )}
          </div>
        )}

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

  useEffect(() => {
    setLocalUnit(unit)
  }, [unit])

  const handleAmountChange = (newAmount: string) => {
    if (newAmount === '') {
      onChange(undefined, localUnit)
      return
    }
    const n = parseSafeInteger(newAmount)
    onChange(n === undefined || n < 0 ? undefined : n, localUnit)
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
          step={1}
          inputMode={'numeric'}
          className="w-20"
          value={amount ?? ''}
          onChange={(e) => handleAmountChange(e.target.value)}
          placeholder="0"
        />
        <Select
          value={localUnit}
          onValueChange={(v) => handleUnitChange(v as DateUnit)}
        >
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
