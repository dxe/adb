import { KeyboardEvent } from 'react'
import { Input } from '@/components/ui/input'
import { cn } from '@/lib/utils'
import { MailX, PhoneMissed, UserRoundPlus, Check } from 'lucide-react'
import { AnyFieldApi } from '@tanstack/react-form'
import { ActivistRegistry } from './activist-registry'
import { Popover, PopoverAnchor } from '@/components/ui/popover'
import { useSuggestions } from './use-suggestions'
import { SuggestionList } from './suggestion-list'

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
  const { suggestions, selectedIndex, onInputChange, onSelect, onKeyDown } =
    useSuggestions((v) => registry.getSuggestions(v))

  // Derive popover open state from suggestions and focus
  const open = suggestions.length > 0 && isFocused

  const handleInputChange = (value: string) => {
    field.handleChange(value)
    onInputChange(value)
    onChange()
  }

  const handleSelectSuggestion = (value: string) => {
    field.handleChange(value)
    field.handleBlur()
    field.validate('change')
    onSelect()
    onAdvanceFocus()
  }

  const handleKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    onKeyDown(e)
    switch (e.key) {
      case 'Enter': {
        e.preventDefault()
        const trimmedValue: string = field.state.value?.trim() ?? ''
        if (!trimmedValue.length) return
        const selectedValue =
          selectedIndex >= 0 && selectedIndex < suggestions.length
            ? suggestions[selectedIndex]
            : trimmedValue
        handleSelectSuggestion(selectedValue)
        return
      }
      case 'Tab': {
        if (e.shiftKey) return
        const trimmedValue = field.state.value?.trim() ?? ''
        if (!trimmedValue.length) return
        e.preventDefault()
        const selectedValue =
          selectedIndex >= 0 && selectedIndex < suggestions.length
            ? suggestions[selectedIndex]
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
          <Popover open={open}>
            <PopoverAnchor asChild>
              <div className="relative">
                <Input
                  ref={inputRef}
                  value={field.state.value ?? ''}
                  onChange={(e) => handleInputChange(e.target.value)}
                  onKeyDown={handleKeyDown}
                  name="attendee"
                  placeholder=""
                  className={cn(
                    'w-full transition-colors duration-300',
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
                    onSelect()
                  }}
                />
                <div className="right-0 top-0 bottom-0 h-full pointer-events-none absolute flex gap-2 items-center p-1.5 opacity-80">
                  {hasAllInfo && <Check className="text-green-500" />}
                  {isNewName && <UserRoundPlus className="text-purple-500" />}
                  {isMissingEmail && <MailX className="text-orange-500" />}
                  {isMissingPhone && (
                    <PhoneMissed className="text-orange-500" />
                  )}
                </div>
              </div>
            </PopoverAnchor>
            <SuggestionList
              suggestions={suggestions}
              selectedIndex={selectedIndex}
              onSelect={handleSelectSuggestion}
            />
          </Popover>
        </div>
      </div>
      {field.state.meta.errors[0] && (
        <p className="text-sm text-red-500 mt-1">
          {field.state.meta.errors[0]?.message}
        </p>
      )}
      {isDuplicate && <p className="text-xs text-red-500">Duplicate entry</p>}
    </div>
  )
}
