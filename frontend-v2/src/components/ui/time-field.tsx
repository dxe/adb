'use client'

import {
  TimeField as AriaTimeField,
  DateInput,
  DateSegment,
  type TimeValue,
} from 'react-aria-components'
import { Time, parseTime } from '@internationalized/date'
import { X } from 'lucide-react'
import { cn } from '@/lib/utils'

export interface TimeFieldProps {
  // 24h "HH:MM" (what the backend stores), or '' when unset.
  value: string
  onChange: (value: string) => void
  // When provided, a clear (✕) button shows inside the field once it has a
  // value, right-aligned against the left-aligned time.
  onClear?: () => void
  className?: string
  hasError?: boolean
  'aria-label'?: string
}

// Parse a stored "HH:MM" (or "HH:MM:SS") string into a Time, tolerating empty
// or malformed input by returning null (which shows the placeholder).
function toTime(value: string): Time | null {
  if (!value) return null
  try {
    return parseTime(value.length > 5 ? value.slice(0, 5) : value)
  } catch {
    return null
  }
}

function toHHMM(time: TimeValue | null): string {
  if (!time) return ''
  return `${String(time.hour).padStart(2, '0')}:${String(time.minute).padStart(2, '0')}`
}

// Accessible, fully-controlled time entry built on react-aria. Used instead of
// native <input type="time">, which is awkward to clear and (in desktop Safari)
// paints a phantom "12:30 PM" on empty fields.
export function TimeField({
  value,
  onChange,
  onClear,
  className,
  hasError,
  'aria-label': ariaLabel,
}: TimeFieldProps) {
  const showClear = Boolean(value && onClear)
  return (
    <AriaTimeField
      value={toTime(value)}
      onChange={(time) => onChange(toHHMM(time))}
      aria-label={ariaLabel}
      className={cn('relative w-full', className)}
    >
      <DateInput
        className={cn(
          'flex h-9 w-full items-center rounded-md border border-input bg-transparent px-3 py-1 text-sm transition-colors hover:border-gray-400 focus-within:border-primary focus-within:ring-1 focus-within:ring-ring',
          showClear && 'pr-8',
          hasError && 'border-red-500',
        )}
      >
        {(segment) => (
          <DateSegment
            segment={segment}
            className={({ isFocused, isPlaceholder }) =>
              cn(
                'rounded outline-none',
                segment.type === 'literal'
                  ? 'text-muted-foreground'
                  : 'px-0.5 tabular-nums',
                isFocused && 'bg-primary text-primary-foreground',
                isPlaceholder && !isFocused && 'text-muted-foreground',
              )
            }
          />
        )}
      </DateInput>
      {showClear && (
        <button
          type="button"
          onClick={onClear}
          aria-label="Clear time"
          className="absolute right-1.5 top-1/2 grid h-5 w-5 -translate-y-1/2 place-content-center rounded text-muted-foreground hover:text-foreground"
        >
          <X className="h-4 w-4" />
        </button>
      )}
    </AriaTimeField>
  )
}
