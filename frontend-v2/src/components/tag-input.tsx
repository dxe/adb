'use client'

import { useId, useMemo, useState } from 'react'
import { X } from 'lucide-react'
import { Label } from '@/components/ui/label'
import { Popover, PopoverAnchor, PopoverContent } from '@/components/ui/popover'
import { cn } from '@/lib/utils'

export interface TagInputProps {
  /**
   * Currently selected values, rendered as removable chips. Order is
   * preserved (new selections are appended).
   */
  value: string[]
  /** Called with the full next array whenever a chip is added or removed. */
  onChange: (next: string[]) => void
  /**
   * The full universe of selectable values (e.g. all activist names). The
   * caller owns fetching/loading this list — this component does not fetch
   * data itself, it only filters/renders what it's given.
   *
   * Values already present in `value` are excluded from suggestions.
   */
  options: string[]
  /**
   * Optional label rendered above the control via the shared `Label`
   * primitive, wired up with `htmlFor`/`id`. Omit this if the caller wants
   * to render its own external `<Label htmlFor={id}>` next to the control
   * (pass `id` in that case so the label's `htmlFor` still resolves).
   */
  label?: string
  /**
   * `id` applied to the text input. Auto-generated if omitted. Only needed
   * explicitly when an external label (not the built-in `label` prop) needs
   * to reference this input via `htmlFor`. Note that the input is unmounted
   * while `disabled` or once `max` selections are reached, so an external
   * label's `htmlFor` dangles in those states — prefer the built-in `label`
   * prop, which handles this.
   */
  id?: string
  /** Placeholder shown in the text input while no chips are selected. */
  placeholder?: string
  /**
   * Maximum number of selected values. Once reached, the text input is
   * hidden (existing chips remain visible and, unless `disabled`,
   * removable). Omit for unbounded selection. Pass `1` for single-select
   * (point person, host, etc.) usage.
   */
  max?: number
  /**
   * Maximum number of matching suggestions shown in the dropdown at once.
   * @default 20
   */
  maxSuggestions?: number
  /** Disables the control: hides the text input and chip-remove buttons. */
  disabled?: boolean
}

/**
 * Chip/tag multi-select with type-ahead filtering over a fixed list of
 * `options`, restricted to those options (there is no free-text/allow-new
 * mode — this matches the legacy Vue `b-taginput` usages this component
 * replaces, which all set `allow-new: false`).
 *
 * Filtering is case-insensitive "starts with", matching the legacy Vue
 * `getFilteredActivists`/`getFilteredOrganizers` behavior (`WorkingGroupList.vue`,
 * `CirclesList.vue`) that both the Working Groups and Circles pages ported from.
 *
 * Keyboard behavior:
 * - `Enter` selects the first suggestion (no-op if there are none).
 * - `Backspace` on an empty input removes the last chip.
 *
 * This is a presentational component: callers own data fetching (typically
 * via a `useQuery` call that loads a plain name list once) and pass the
 * result in as `options`.
 */
export function TagInput({
  value,
  onChange,
  options,
  label,
  id,
  placeholder = 'Search by name...',
  max,
  maxSuggestions = 20,
  disabled = false,
}: TagInputProps) {
  const [text, setText] = useState('')
  const [open, setOpen] = useState(false)
  const generatedId = useId()
  const inputId = id ?? generatedId
  const listboxId = `${inputId}-listbox`

  const atMax = max !== undefined && value.length >= max
  const showInput = !disabled && !atMax

  const suggestions = useMemo(() => {
    const query = text.trim().toLowerCase()
    if (!query) return []
    return options
      .filter(
        (name) => !value.includes(name) && name.toLowerCase().startsWith(query),
      )
      .slice(0, maxSuggestions)
  }, [options, text, value, maxSuggestions])

  const addValue = (name: string) => {
    if (!name.trim() || value.includes(name) || atMax) return
    onChange([...value, name])
    setText('')
    setOpen(false)
  }

  const removeValue = (name: string) => {
    onChange(value.filter((v) => v !== name))
  }

  const dropdownOpen = open && suggestions.length > 0

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
        <Popover open={dropdownOpen} onOpenChange={setOpen}>
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
              id={listboxId}
              role="listbox"
              className="max-h-[240px] overflow-y-auto rounded-md border border-gray-200 bg-white shadow-lg"
            >
              {suggestions.map((s, i) => (
                <li
                  key={s}
                  id={`${listboxId}-option-${i}`}
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
      {/* Only reference the input while it's actually rendered — it's hidden
          when `disabled` or once `max` selections are reached, and a label
          must not point at a non-existent element. */}
      <Label htmlFor={showInput ? inputId : undefined}>{label}</Label>
      {control}
    </div>
  )
}
