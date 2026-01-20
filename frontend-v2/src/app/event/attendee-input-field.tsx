import { KeyboardEvent, useState } from 'react'
import { Input } from '@/components/ui/input'
import { cn } from '@/lib/utils'
import { MailX, PhoneMissed, UserRoundPlus, Check } from 'lucide-react'
import { AnyFieldApi } from '@tanstack/react-form'
import { ActivistRegistry } from './activist-registry'

type AttendeeInputFieldProps = {
  field: AnyFieldApi
  index: number
  isFocused: boolean
  inputRef: (el: HTMLInputElement | null) => void
  onFocus: (index: number) => void
  onAdvanceFocus: () => void
  onChange: () => void
  registry: ActivistRegistry
  checkForDuplicate: (value: string, index: number) => boolean
}

export const AttendeeInputField = ({
  field,
  index,
  isFocused,
  inputRef,
  onFocus,
  onAdvanceFocus,
  onChange,
  registry,
  checkForDuplicate,
}: AttendeeInputFieldProps) => {
  const [suggestions, setSuggestions] = useState<string[]>([])
  const [selectedSuggestionIndex, setSelectedSuggestionIndex] = useState(-1)

  const handleInputChange = (value: string) => {
    field.handleChange(value)
    setSuggestions(registry.getSuggestions(value))
    setSelectedSuggestionIndex(-1)
    onChange()
  }

  const handleSelectSuggestion = (value: string) => {
    field.handleChange(value)
    field.handleBlur()
    field.validate('change')
    setSuggestions([])
    setSelectedSuggestionIndex(-1)
    onAdvanceFocus()
  }

  const handleKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    switch (e.key) {
      case 'ArrowDown': {
        e.preventDefault()
        setSelectedSuggestionIndex((prev) =>
          prev === suggestions.length - 1 ? 0 : prev + 1,
        )
        return
      }
      case 'ArrowUp': {
        e.preventDefault()
        setSelectedSuggestionIndex((prev) =>
          prev === 0 ? suggestions.length - 1 : prev - 1,
        )
        return
      }
      case 'Escape': {
        setSuggestions([])
        return
      }
      case 'Enter': {
        e.preventDefault()
        const trimmedValue: string = field.state.value?.trim() ?? ''
        if (!trimmedValue.length) {
          return
        }
        const selectedValue =
          selectedSuggestionIndex >= 0
            ? suggestions[selectedSuggestionIndex]
            : trimmedValue
        handleSelectSuggestion(selectedValue)
        return
      }
      case 'Tab': {
        if (e.shiftKey) {
          return
        }
        const trimmedValue = field.state.value?.trim() ?? ''
        if (!trimmedValue.length) {
          return
        }
        e.preventDefault()
        const selectedValue =
          selectedSuggestionIndex >= 0
            ? suggestions[selectedSuggestionIndex]
            : trimmedValue
        handleSelectSuggestion(selectedValue)
      }
    }
  }

  const trimmedName = field.state.value?.trim() ?? ''
  const isDuplicate = !!trimmedName && checkForDuplicate(trimmedName, index)
  const activist = registry.getActivist(trimmedName)
  const isNewName = !!trimmedName && !activist
  const isExisting = !!trimmedName && !!activist
  const isMissingEmail = isExisting && !activist.email
  const isMissingPhone = isExisting && !activist.phone
  const hasAllInfo = isExisting && !isMissingEmail && !isMissingPhone
  const isError = !!field.state.meta.errors[0]

  return (
    <div className="relative">
      <div className="flex items-center gap-2">
        <div className="relative w-full">
          <div className="relative">
            <Input
              ref={inputRef}
              value={field.state.value ?? ''}
              onChange={(e) => handleInputChange(e.target.value)}
              onKeyDown={handleKeyDown}
              name="attendee"
              placeholder=""
              className={cn(
                'w-full transition-colors duration-300 border-2',
                isDuplicate || isError
                  ? 'text-red-500 border-red-500 focus:border-red-500'
                  : isNewName
                    ? 'border-purple-500 focus:border-transparent'
                    : '',
              )}
              autoComplete="off"
              onFocus={() => onFocus(index)}
              onBlur={() => {
                field.handleBlur()
                setSuggestions([])
              }}
            />
            <div className="right-0 top-0 bottom-0 h-full pointer-events-none absolute flex gap-2 items-center p-1.5 opacity-80">
              {hasAllInfo && <Check className="text-green-500" />}
              {isNewName && <UserRoundPlus className="text-purple-500" />}
              {isMissingEmail && <MailX className="text-orange-500" />}
              {isMissingPhone && <PhoneMissed className="text-orange-500" />}
            </div>
          </div>
          {isFocused && !!suggestions.length && (
            <ul className="absolute z-10 mt-1 w-full rounded-md border border-gray-200 bg-white shadow-lg">
              {suggestions.map((suggestion, i) => (
                <li
                  key={suggestion}
                  className={cn(
                    'cursor-pointer px-4 py-2 hover:bg-gray-100',
                    i === selectedSuggestionIndex ? 'bg-neutral-100' : '',
                  )}
                  onMouseDown={(e) => {
                    // Use onMouseDown instead of onClick to fire before onBlur.
                    e.preventDefault()
                    handleSelectSuggestion(suggestion)
                  }}
                >
                  {suggestion}
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>
      {field.state.meta.errors[0] && (
        <p className="text-sm text-red-500 mt-1">
          {field.state.meta.errors[0]?.message}
        </p>
      )}
      {isDuplicate && (
        <p className="text-sm text-red-500 mt-1">Duplicate entry</p>
      )}
    </div>
  )
}
