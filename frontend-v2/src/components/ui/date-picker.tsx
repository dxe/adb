'use client'

import * as React from 'react'
import { Calendar } from '@/components/ui/calendar'
import { Input } from '@/components/ui/input'
import { Popover, PopoverAnchor, PopoverContent } from '@/components/ui/popover'
import { format, isValid, parse } from 'date-fns'
import { CalendarIcon } from 'lucide-react'
import { cn } from '@/lib/utils'

export interface DatePickerProps {
  value?: Date
  onValueChange?: (date: Date | undefined) => void
  placeholder?: string
  className?: string
  disabled?: boolean
}

const PARSE_FORMATS = [
  'MM/dd/yyyy',
  'M/d/yyyy',
  'MM-dd-yyyy',
  'M-d-yyyy',
  'yyyy-MM-dd',
]

function parseDate(str: string): Date | undefined {
  for (const fmt of PARSE_FORMATS) {
    const parsed = parse(str, fmt, new Date())
    if (isValid(parsed)) return parsed
  }
  return undefined
}

const DISPLAY_FORMAT = 'MM/dd/yyyy'

// Fixed 8-slot string; '_' means the slot has not been filled yet.
// Slots: [M0 M1 D0 D1 Y0 Y1 Y2 Y3]
const EMPTY_DIGITS = '________'

function buildMasked(digits: string): string {
  const ph = ['M', 'M', 'D', 'D', 'Y', 'Y', 'Y', 'Y']
  const d = (i: number) => (digits[i] !== '_' ? digits[i] : ph[i])
  return `${d(0)}${d(1)}/${d(2)}${d(3)}/${d(4)}${d(5)}${d(6)}${d(7)}`
}

// Maps digit slot index (0–8) to display cursor position (0–10).
function cursorPos(n: number): number {
  if (n <= 2) return n
  if (n <= 4) return n + 1
  return n + 2
}

// Maps display cursor position (0–10) to digit slot index (0–8).
// Display: M M / D D / Y Y Y Y  (chars 0–9, cursor positions 0–10)
function displayCursorToDigitCursor(pos: number): number {
  if (pos <= 2) return pos
  if (pos <= 5) return pos - 1
  return pos - 2
}

function extractDigits(str: string): string {
  return str.replace(/\D/g, '').slice(0, 8)
}

function fromDate(date: Date): string {
  return extractDigits(format(date, DISPLAY_FORMAT))
}

export function DatePicker({
  value,
  onValueChange,
  placeholder = 'MM/DD/YYYY',
  className,
  disabled,
}: DatePickerProps) {
  const [open, setOpen] = React.useState(false)
  const [digits, setDigits] = React.useState(
    value ? fromDate(value) : EMPTY_DIGITS,
  )
  const [calendarMonth, setCalendarMonth] = React.useState<Date | undefined>(
    value,
  )
  const [isFocused, setIsFocused] = React.useState(false)
  const inputRef = React.useRef<HTMLInputElement>(null)
  const containerRef = React.useRef<HTMLDivElement>(null)
  const skipNextOpenRef = React.useRef(false)
  const pendingCursorRef = React.useRef<number | null>(null)

  // Sync when value changes externally (e.g. parent resets form).
  React.useEffect(() => {
    if (!isFocused) {
      setDigits(value ? fromDate(value) : EMPTY_DIGITS)
      if (value) setCalendarMonth(value)
    }
  }, [value, isFocused])

  // Apply pending cursor position after DOM update, before paint.
  React.useLayoutEffect(() => {
    if (pendingCursorRef.current !== null) {
      inputRef.current?.setSelectionRange(
        pendingCursorRef.current,
        pendingCursorRef.current,
      )
      pendingCursorRef.current = null
    }
  })

  const displayValue =
    isFocused || digits !== EMPTY_DIGITS ? buildMasked(digits) : ''

  function moveCursor(n: number) {
    pendingCursorRef.current = cursorPos(n)
  }

  function isAllSelected(): boolean {
    const el = inputRef.current
    return (
      !!el && el.selectionStart === 0 && el.selectionEnd === el.value.length
    )
  }

  function getSelectionDigitRange(): { dStart: number; dEnd: number } | null {
    const el = inputRef.current
    if (!el) return null
    const s = el.selectionStart ?? 8
    const e = el.selectionEnd ?? 8
    if (s === e) return null
    return {
      dStart: displayCursorToDigitCursor(s),
      dEnd: displayCursorToDigitCursor(e),
    }
  }

  function commitDigits(d: string) {
    setDigits(d)
    if (!d.includes('_')) {
      const parsed = parseDate(buildMasked(d))
      if (parsed) {
        onValueChange?.(parsed)
        setCalendarMonth(parsed)
      }
    } else if (d === EMPTY_DIGITS) {
      onValueChange?.(undefined)
    }
  }

  function resetToLastValid() {
    setDigits(value ? fromDate(value) : EMPTY_DIGITS)
  }

  function closeAndReset() {
    setOpen(false)
    const isComplete = !digits.includes('_') && !!parseDate(buildMasked(digits))
    const isEmpty = digits === EMPTY_DIGITS
    if (!isComplete && !isEmpty) resetToLastValid()
  }

  function handleKeyDown(e: React.KeyboardEvent<HTMLInputElement>) {
    if (e.ctrlKey || e.metaKey) return

    if (e.key === 'Enter') {
      if (open) {
        e.preventDefault()
        e.stopPropagation()
        closeAndReset()
      }
      return
    }

    if (e.key >= '0' && e.key <= '9') {
      e.preventDefault()
      if (isAllSelected()) {
        commitDigits(e.key + EMPTY_DIGITS.slice(1))
        moveCursor(1)
      } else {
        const sel = getSelectionDigitRange()
        if (sel) {
          // Fill from the start of the selection, clear the rest of the
          // selected range so adjacent segments don't bleed in.
          const segLen = sel.dEnd - sel.dStart
          const next =
            digits.slice(0, sel.dStart) +
            e.key +
            '_'.repeat(segLen - 1) +
            digits.slice(sel.dEnd)
          commitDigits(next)
          moveCursor(sel.dStart + 1)
        } else {
          // Overwrite at cursor: replaces the slot under the cursor.
          const displayPos = inputRef.current?.selectionStart ?? cursorPos(8)
          const dPos = displayCursorToDigitCursor(displayPos)
          if (dPos < 8) {
            const next = digits.slice(0, dPos) + e.key + digits.slice(dPos + 1)
            commitDigits(next)
            moveCursor(dPos + 1)
          }
        }
      }
    } else if (e.key === 'Backspace') {
      e.preventDefault()
      if (isAllSelected()) {
        commitDigits(EMPTY_DIGITS)
        moveCursor(0)
      } else {
        const sel = getSelectionDigitRange()
        if (sel) {
          const segLen = sel.dEnd - sel.dStart
          const next =
            digits.slice(0, sel.dStart) +
            '_'.repeat(segLen) +
            digits.slice(sel.dEnd)
          commitDigits(next)
          moveCursor(sel.dStart)
        } else {
          // Clear the slot just before the cursor.
          const displayPos = inputRef.current?.selectionStart ?? 0
          const dPos = displayCursorToDigitCursor(displayPos)
          if (dPos > 0) {
            const target = dPos - 1
            const next =
              digits.slice(0, target) + '_' + digits.slice(target + 1)
            commitDigits(next)
            moveCursor(target)
          }
        }
      }
    } else if (e.key === 'Delete') {
      e.preventDefault()
      commitDigits(EMPTY_DIGITS)
      moveCursor(0)
    } else if (e.key.length === 1) {
      e.preventDefault()
    }
  }

  // Handles paste via onChange (keydown e.preventDefault suppresses normal input).
  function handleChange(e: React.ChangeEvent<HTMLInputElement>) {
    const raw = extractDigits(e.target.value)
    const padded = (raw + EMPTY_DIGITS).slice(0, 8)
    commitDigits(padded)
    moveCursor(Math.min(raw.length, 8))
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverAnchor asChild>
        <div ref={containerRef} className={cn('relative', className)}>
          <Input
            ref={inputRef}
            value={displayValue}
            onChange={handleChange}
            onKeyDown={handleKeyDown}
            onClick={() => {
              if (!disabled) setOpen(true)
            }}
            onFocus={() => {
              setIsFocused(true)
              const n = digits.indexOf('_')
              moveCursor(n === -1 ? 8 : n)
              if (skipNextOpenRef.current) {
                skipNextOpenRef.current = false
                return
              }
              if (!disabled) setOpen(true)
            }}
            onBlur={() => {
              setIsFocused(false)
              // Reset partial/invalid input when calendar is also closed.
              if (!open) {
                const isComplete =
                  !digits.includes('_') && !!parseDate(buildMasked(digits))
                const isEmpty = digits === EMPTY_DIGITS
                if (!isComplete && !isEmpty) resetToLastValid()
              }
            }}
            disabled={disabled}
            placeholder={placeholder}
            style={{ paddingLeft: '2.5rem' }}
          />
          <button
            type="button"
            disabled={disabled}
            onMouseDown={(e) => e.preventDefault()}
            onClick={() => {
              if (!disabled) {
                setOpen((prev) => !prev)
                skipNextOpenRef.current = true
                inputRef.current?.focus()
              }
            }}
            className="absolute inset-y-0 left-3 flex items-center text-muted-foreground hover:text-foreground disabled:opacity-50"
            tabIndex={-1}
          >
            <CalendarIcon className="h-4 w-4" />
          </button>
        </div>
      </PopoverAnchor>
      <PopoverContent
        className="w-auto p-0"
        align="start"
        onOpenAutoFocus={(e) => e.preventDefault()}
        onMouseDown={(e) => e.preventDefault()}
        onInteractOutside={(e) => {
          // Don't dismiss when interacting with the input/icon — we manage that ourselves.
          if (containerRef.current?.contains(e.target as Node)) {
            e.preventDefault()
          }
        }}
        onFocusOutside={(e) => {
          if (containerRef.current?.contains(e.target as Node)) {
            e.preventDefault()
          }
        }}
        onEscapeKeyDown={() => closeAndReset()}
      >
        <Calendar
          mode="single"
          selected={value}
          month={calendarMonth}
          onMonthChange={setCalendarMonth}
          onSelect={(date) => {
            onValueChange?.(date)
            setDigits(date ? fromDate(date) : EMPTY_DIGITS)
            if (date) setCalendarMonth(date)
            setOpen(false)
            skipNextOpenRef.current = true
            inputRef.current?.focus()
          }}
        />
      </PopoverContent>
    </Popover>
  )
}
