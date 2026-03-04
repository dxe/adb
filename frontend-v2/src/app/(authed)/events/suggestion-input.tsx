import { ComponentProps, KeyboardEvent, ReactNode, useState } from 'react'
import { Input } from '@/components/ui/input'
import { Popover, PopoverAnchor } from '@/components/ui/popover'
import { SuggestionList } from './suggestion-list'

type SelectMeta = {
  key: 'Enter' | 'Tab' | 'click'
  fromSuggestion: boolean
}

type Props = {
  value: string
  onValueChange: (value: string) => void
  getSuggestions: (value: string) => string[]
  onCommit?: (meta: SelectMeta) => void
  isFocused?: boolean
  inputRef?: (el: HTMLInputElement | null) => void
  size?: 'sm' | 'base'
  endAdornment?: ReactNode
  onFocus?: () => void
  onBlur?: () => void
  onKeyDown?: (event: KeyboardEvent<HTMLInputElement>) => void
} & Omit<
  ComponentProps<typeof Input>,
  'value' | 'onChange' | 'onKeyDown' | 'onFocus' | 'onBlur' | 'ref' | 'size'
>

export function SuggestionInput({
  value,
  onValueChange,
  getSuggestions,
  onCommit,
  isFocused,
  inputRef,
  size = 'base',
  endAdornment,
  onFocus,
  onBlur,
  onKeyDown,
  autoComplete = 'off',
  ...inputProps
}: Props) {
  const [selection, setSelection] = useState<{ index: number; value: string }>({
    index: -1,
    value: '',
  })
  const [focused, setFocused] = useState(false)
  const [isOpen, setIsOpen] = useState(false)
  const hasFocus = isFocused ?? focused
  const shouldFetchSuggestions = isOpen && hasFocus && value.trim().length > 0
  const suggestions = shouldFetchSuggestions ? getSuggestions(value) : []
  const selectedIndex = selection.value === value ? selection.index : -1
  const open = shouldFetchSuggestions && suggestions.length > 0

  const clearSuggestions = () => {
    setSelection({ index: -1, value })
  }

  const handleChange = (nextValue: string) => {
    setIsOpen(true)
    setSelection({ index: -1, value: nextValue })
    onValueChange(nextValue)
  }

  const handleSuggestionSelect = (suggestion: string) => {
    setIsOpen(false)
    clearSuggestions()
    onValueChange(suggestion)
    onCommit?.({ key: 'click', fromSuggestion: true })
  }

  const handleKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
    onKeyDown?.(event)
    if (event.defaultPrevented) return

    if (event.key === 'ArrowDown') {
      if (!suggestions.length) return
      event.preventDefault()
      setSelection((prev) => {
        const currentIndex = prev.value === value ? prev.index : -1
        const nextIndex =
          currentIndex === suggestions.length - 1 ? 0 : currentIndex + 1
        return { index: nextIndex, value }
      })
      return
    }
    if (event.key === 'ArrowUp') {
      if (!suggestions.length) return
      event.preventDefault()
      setSelection((prev) => {
        const currentIndex = prev.value === value ? prev.index : -1
        const nextIndex =
          currentIndex <= 0 ? suggestions.length - 1 : currentIndex - 1
        return { index: nextIndex, value }
      })
      return
    }
    if (event.key === 'Escape') {
      setIsOpen(false)
      clearSuggestions()
      return
    }

    const isTabCommit = event.key === 'Tab' && !event.shiftKey
    const shouldCommit = event.key === 'Enter' || isTabCommit
    if (!shouldCommit) return
    if (isTabCommit && !open) return

    const selectedSuggestion =
      selectedIndex >= 0 && selectedIndex < suggestions.length
        ? suggestions[selectedIndex]
        : null
    const resolvedValue = selectedSuggestion ?? value.trim()

    event.preventDefault()
    onValueChange(resolvedValue)
    onCommit?.({
      key: event.key as 'Enter' | 'Tab',
      fromSuggestion: !!selectedSuggestion,
    })
    setIsOpen(false)
    clearSuggestions()
  }

  const handleFocus = () => {
    setFocused(true)
    setIsOpen(true)
    onFocus?.()
  }

  const handleBlur = () => {
    setFocused(false)
    setIsOpen(false)
    clearSuggestions()
    onBlur?.()
  }

  const listboxId = `${inputProps.id ?? 'suggestion-input'}-listbox`
  const activeOptionId =
    selectedIndex >= 0 ? `${listboxId}-option-${selectedIndex}` : undefined

  return (
    <Popover
      open={open}
      onOpenChange={(nextOpen) => {
        setIsOpen(nextOpen)
      }}
    >
      <PopoverAnchor asChild>
        <div className="relative">
          <Input
            role="combobox"
            aria-autocomplete="list"
            aria-expanded={open}
            aria-controls={listboxId}
            aria-activedescendant={activeOptionId}
            ref={inputRef}
            value={value}
            onChange={(e) => handleChange(e.target.value)}
            onKeyDown={handleKeyDown}
            onFocus={handleFocus}
            onBlur={handleBlur}
            autoComplete={autoComplete}
            {...inputProps}
          />
          {endAdornment}
        </div>
      </PopoverAnchor>
      <SuggestionList
        listboxId={listboxId}
        suggestions={suggestions}
        selectedIndex={selectedIndex}
        onSelect={handleSuggestionSelect}
        size={size}
      />
    </Popover>
  )
}
