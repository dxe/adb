'use client'

import { useId, useMemo, useState } from 'react'
import { X } from 'lucide-react'
import { Label } from '@/components/ui/label'
import { Popover, PopoverAnchor, PopoverContent } from '@/components/ui/popover'
import { cn } from '@/lib/utils'

export interface TagInputProps {
  /** Selected values, rendered as removable chips (new selections are appended). */
  value: string[]
  /** Called with the full next array whenever a chip is added or removed. */
  onChange: (next: string[]) => void
  /** All selectable values — callers own data fetching; already-selected values are excluded from suggestions. */
  options: string[]
  /** Optional label rendered above the control, wired to the input via `htmlFor`. */
  label?: string
  /** Placeholder shown in the text input while no chips are selected. */
  placeholder?: string
  /** Allow only one selection — the input hides while a value is picked. */
  single?: boolean
  /** Max suggestions shown in the dropdown (default 20). */
  maxSuggestions?: number
  /** Disables the control: hides the text input and chip-remove buttons. */
  disabled?: boolean
}

/**
 * Chip multi-select with type-ahead restricted to `options`, matched by case-insensitive
 * prefix to mirror the legacy Vue `b-taginput` (allow-new: false) usages it replaces.
 */
export function TagInput({
  value,
  onChange,
  options,
  label,
  placeholder = 'Search by name...',
  single = false,
  maxSuggestions = 20,
  disabled = false,
}: TagInputProps) {
  const [text, setText] = useState('')
  const [isOpen, setIsOpen] = useState(false)
  const inputId = useId()
  const listboxId = `${inputId}-listbox`

  const atLimit = single && value.length > 0
  const showInput = !disabled && !atLimit
  // The input unmounts while hidden, so the label must not reference it then.
  const labelFor = showInput ? inputId : undefined

  const selectedSet = useMemo(() => new Set(value), [value])
  const suggestions = useMemo(() => {
    const query = text.trim().toLowerCase()
    if (!query) return []
    return options
      .filter(
        (name) =>
          !selectedSet.has(name) && name.toLowerCase().startsWith(query),
      )
      .slice(0, maxSuggestions)
  }, [options, text, selectedSet, maxSuggestions])

  const addValue = (name: string) => {
    if (!name.trim() || selectedSet.has(name) || atLimit) return
    onChange([...value, name])
    setText('')
    setIsOpen(false)
  }

  const removeValue = (name: string) => {
    onChange(value.filter((v) => v !== name))
  }

  const dropdownOpen = isOpen && suggestions.length > 0

  const control = (
    <div
      className={cn(
        'flex min-h-9 flex-wrap items-center gap-1.5 rounded-md border border-input bg-transparent px-2 py-1 text-sm',
        'focus-within:border-primary focus-within:ring-1 focus-within:ring-ring',
        disabled && 'cursor-not-allowed opacity-50',
      )}
    >
      {value.map((name) => (
        <span
          key={name}
          className="inline-flex items-center gap-1 rounded-full bg-secondary px-2 py-0.5 text-xs text-secondary-foreground"
        >
          {name}
          {!disabled && (
            <button
              type="button"
              onClick={() => removeValue(name)}
              aria-label={`Remove ${name}`}
              className="rounded-full hover:bg-secondary-foreground/10"
            >
              <X className="h-3 w-3" />
            </button>
          )}
        </span>
      ))}
      {showInput && (
        <Popover open={dropdownOpen} onOpenChange={setIsOpen}>
          <PopoverAnchor asChild>
            <input
              id={inputId}
              role="combobox"
              aria-autocomplete="list"
              aria-expanded={dropdownOpen}
              aria-controls={listboxId}
              className="min-w-[8rem] flex-1 border-0 bg-transparent p-1 text-sm outline-none placeholder:text-muted-foreground"
              value={text}
              placeholder={value.length === 0 ? placeholder : undefined}
              onChange={(e) => {
                setText(e.target.value)
                setIsOpen(true)
              }}
              onFocus={() => setIsOpen(true)}
              onBlur={() => setIsOpen(false)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') {
                  e.preventDefault()
                  if (suggestions.length > 0) addValue(suggestions[0])
                } else if (
                  e.key === 'Backspace' &&
                  text === '' &&
                  value.length > 0
                ) {
                  removeValue(value[value.length - 1])
                }
              }}
            />
          </PopoverAnchor>
          <PopoverContent
            className="w-[var(--radix-popover-trigger-width)] p-0"
            align="start"
            sideOffset={4}
            onOpenAutoFocus={(e) => e.preventDefault()}
            onCloseAutoFocus={(e) => e.preventDefault()}
          >
            <ul
              id={listboxId}
              role="listbox"
              className="max-h-[240px] overflow-y-auto rounded-md border border-gray-200 bg-white shadow-lg"
            >
              {suggestions.map((s) => (
                <li
                  key={s}
                  role="option"
                  aria-selected={false}
                  className="cursor-pointer px-3 py-1 text-sm hover:bg-gray-100"
                  onMouseDown={(e) => {
                    // Fires before the input's onBlur closes the popover.
                    e.preventDefault()
                    addValue(s)
                  }}
                >
                  {s}
                </li>
              ))}
            </ul>
          </PopoverContent>
        </Popover>
      )}
    </div>
  )

  if (!label) return control

  return (
    <div className="space-y-1.5">
      <Label htmlFor={labelFor}>{label}</Label>
      {control}
    </div>
  )
}
