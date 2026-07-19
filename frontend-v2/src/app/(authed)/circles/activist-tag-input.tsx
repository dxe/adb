'use client'

import { useMemo, useState } from 'react'
import { X } from 'lucide-react'
import { Popover, PopoverAnchor, PopoverContent } from '@/components/ui/popover'
import { cn } from '@/lib/utils'

type Props = {
  value: string[]
  onChange: (value: string[]) => void
  options: string[]
  placeholder?: string
  /** Once `value.length` reaches this, no more items can be added (matches
   * Buefy's `b-taginput maxtags` behavior used by the legacy host field). */
  maxItems?: number
  disabled?: boolean
  id?: string
}

/**
 * A chip/tag multi-select with autocomplete, filtered by "starts with" like
 * the legacy Vue page's `getFilteredActivists`. There's no existing
 * Combobox/MultiSelect/TagInput primitive in `components/ui`, so this is a
 * small purpose-built component using the existing Popover primitive.
 */
export function ActivistTagInput({
  value,
  onChange,
  options,
  placeholder,
  maxItems,
  disabled,
  id,
}: Props) {
  const [text, setText] = useState('')
  const [open, setOpen] = useState(false)

  const atMax = maxItems !== undefined && value.length >= maxItems

  const suggestions = useMemo(() => {
    if (!text.trim()) return []
    const lower = text.toLowerCase()
    return options
      .filter(
        (name) => !value.includes(name) && name.toLowerCase().startsWith(lower),
      )
      .slice(0, 20)
  }, [options, text, value])

  const addValue = (name: string) => {
    if (!name.trim() || value.includes(name) || atMax) return
    onChange([...value, name])
    setText('')
    setOpen(false)
  }

  const removeValue = (name: string) => {
    onChange(value.filter((v) => v !== name))
  }

  const showInput = !disabled && !atMax

  return (
    <div
      className={cn(
        'flex min-h-9 flex-wrap items-center gap-1 rounded-md border border-input bg-transparent px-2 py-1 text-sm',
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
        <Popover open={open && suggestions.length > 0} onOpenChange={setOpen}>
          <PopoverAnchor asChild>
            <input
              id={id}
              className="min-w-[8rem] flex-1 border-0 bg-transparent p-1 text-sm outline-none placeholder:text-muted-foreground"
              value={text}
              placeholder={value.length === 0 ? placeholder : undefined}
              onChange={(e) => {
                setText(e.target.value)
                setOpen(true)
              }}
              onFocus={() => setOpen(true)}
              onBlur={() => setOpen(false)}
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
}
