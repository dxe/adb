'use client'

import { useMemo, useState } from 'react'
import { X } from 'lucide-react'
import { Label } from '@/components/ui/label'
import { cn } from '@/lib/utils'

/**
 * Chip-based multi-select for picking existing activist/organizer names,
 * with type-ahead filtering. Mirrors the Buefy `b-taginput` used on the
 * legacy Working Groups page (autocomplete, allow-new: false).
 */
export function PersonMultiSelect({
  label,
  options,
  value,
  onChange,
  max,
  placeholder = 'Search by name...',
}: {
  label: string
  options: string[]
  value: string[]
  onChange: (next: string[]) => void
  /** Maximum number of selections. When omitted, selection is unbounded. */
  max?: number
  placeholder?: string
}) {
  const [query, setQuery] = useState('')
  const [isOpen, setIsOpen] = useState(false)

  const selectedSet = useMemo(() => new Set(value), [value])
  const atMax = max !== undefined && value.length >= max

  const suggestions = useMemo(() => {
    const q = query.trim().toLowerCase()
    if (!q) return []
    return options
      .filter((o) => !selectedSet.has(o) && o.toLowerCase().includes(q))
      .slice(0, 8)
  }, [options, query, selectedSet])

  const addValue = (name: string) => {
    if (selectedSet.has(name)) return
    onChange(max === 1 ? [name] : [...value, name])
    setQuery('')
    setIsOpen(false)
  }

  const removeValue = (name: string) => {
    onChange(value.filter((v) => v !== name))
  }

  return (
    <div className="space-y-1.5">
      <Label>{label}</Label>
      <div className="relative">
        <div
          className={cn(
            'flex flex-wrap items-center gap-1.5 rounded-md border border-input bg-transparent p-2',
            'focus-within:border-primary focus-within:ring-1 focus-within:ring-ring',
          )}
        >
          {value.map((name) => (
            <span
              key={name}
              className="inline-flex items-center gap-1 rounded-full bg-secondary px-2 py-0.5 text-xs text-secondary-foreground"
            >
              {name}
              <button
                type="button"
                onClick={() => removeValue(name)}
                aria-label={`Remove ${name}`}
                className="rounded-full hover:opacity-70"
              >
                <X className="h-3 w-3" />
              </button>
            </span>
          ))}
          {!atMax && (
            <input
              className="min-w-[8rem] flex-1 bg-transparent text-sm outline-none placeholder:text-muted-foreground"
              value={query}
              placeholder={value.length === 0 ? placeholder : ''}
              onChange={(e) => {
                setQuery(e.target.value)
                setIsOpen(true)
              }}
              onFocus={() => setIsOpen(true)}
              onBlur={() => setIsOpen(false)}
              onKeyDown={(e) => {
                if (e.key === 'Enter' && suggestions.length > 0) {
                  e.preventDefault()
                  addValue(suggestions[0])
                } else if (
                  e.key === 'Backspace' &&
                  query === '' &&
                  value.length > 0
                ) {
                  removeValue(value[value.length - 1])
                }
              }}
            />
          )}
        </div>
        {isOpen && suggestions.length > 0 && (
          <ul className="absolute z-10 mt-1 max-h-56 w-full overflow-auto rounded-md border bg-popover text-sm shadow-md">
            {suggestions.map((s) => (
              <li key={s}>
                <button
                  type="button"
                  // Fires before the input's onBlur closes the list.
                  onMouseDown={(e) => e.preventDefault()}
                  onClick={() => addValue(s)}
                  className="w-full px-3 py-1.5 text-left hover:bg-accent"
                >
                  {s}
                </button>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  )
}
