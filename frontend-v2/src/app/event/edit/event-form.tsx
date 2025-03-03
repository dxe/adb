'use client'

import { useRef, useState, KeyboardEvent } from 'react'
import { useForm, useFieldArray, Controller } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { MailX, PhoneMissed, UserRoundPlus } from 'lucide-react'

// TODO(jh):
// - fix focus bugs... keyboard nav is working perfectly, but clicking on a suggestion doesn't work.
// - get list of names from server
// - show indicator for missing phone/email
// - add other form fields (event name, type, date, suppress survey)
// - show total # of attendees next to save button
// - handle loading of existing event list (if event id in url param)
// - submit to server on save
// - split into components that make sense
// - show warning if leaving page while form is dirty
// - store list of names from server in indexed db & only update what's been created, updated, or deleted since last load?
// - consider storing unsaved data in session storage to prevent accidental loss?
// - dark mode (nice to have)

// Mock data for suggestions
const people = [
  'John Smith',
  'Jane Doe',
  'Michael Johnson',
  'Emily Davis',
  'Robert Wilson',
  'Sarah Brown',
  'David Miller',
  'Jennifer Taylor',
  'William Anderson',
  'Lisa Thomas',
  'James Jackson',
  'Mary White',
  'Richard Harris',
  'Patricia Martin',
  'Charles Thompson',
]

const DEFAULT_FIELD_COUNT = 5
const MIN_EMPTY_FIELDS = 2

// Zod schema for form validation
const formSchema = z.object({
  attendees: z
    .array(
      z.object({
        // TODO(jh): restrict name length?
        name: z.string().transform((it) => it.trim()),
      }),
    )
    .transform((it) => it.filter(({ name }) => !!name.length)),
})

type FormValues = z.infer<typeof formSchema>

export const EventForm = () => {
  const { control, handleSubmit, watch, setValue } = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      attendees: Array(DEFAULT_FIELD_COUNT).fill({ name: '' }),
    },
  })
  const { fields, append } = useFieldArray({
    control,
    name: 'attendees',
    keyName: '_id',
  })
  const attendees = watch('attendees')

  const inputRefs = useRef<(HTMLInputElement | null)[]>(
    Array(DEFAULT_FIELD_COUNT).fill(null),
  )
  const [suggestions, setSuggestions] = useState<Array<string>>([])
  const [activeInputIndex, setActiveInputIndex] = useState(0)
  const [selectedSuggestionIndex, setSelectedSuggestionIndex] =
    useState<number>(-1)

  const handleInputChange = (index: number, value: string) => {
    const trimmedValue = value.trim()
    if (trimmedValue) {
      // We have a value, so update it in form state.
      setValue(`attendees.${index}.name`, value)
    } else {
      // Clear input if it's nothing but white space.
      setValue(`attendees.${index}.name`, '')
    }

    if (!trimmedValue.length) {
      setSuggestions([])
    } else {
      // TODO(jh): match filtering logic that is currently in the Vue app.
      const filtered = people.filter((person) =>
        person.toLowerCase().includes(value.toLowerCase()),
      )
      setSuggestions(filtered)
    }
    setSelectedSuggestionIndex(-1)
  }

  const handleSelectSuggestion = (index: number, value: string) => {
    setValue(`attendees.${index}.name`, value)
    setSuggestions([])
    if (index < fields.length - 1) {
      inputRefs.current[index + 1]?.focus()
    }
  }

  const handleKeyDown = (e: KeyboardEvent<HTMLInputElement>, index: number) => {
    switch (e.key) {
      case 'ArrowDown': {
        // Move down in the list of suggestions.
        setSelectedSuggestionIndex((prev) =>
          prev === suggestions.length - 1 ? 0 : prev + 1,
        )
        return
      }
      case 'ArrowUp': {
        // Move up in the list of suggestions.
        setSelectedSuggestionIndex((prev) =>
          prev === 0 ? suggestions.length - 1 : prev - 1,
        )
        return
      }
      case 'Escape': {
        // Dismiss the suggestions.
        setSuggestions([])
        return
      }
      case 'Enter': {
        // Behave similarly to Tab.
        e.preventDefault()
        const trimmedValue = attendees[index].name.trim()
        if (!trimmedValue.length) {
          return
        }
        handleSelectSuggestion(
          index,
          selectedSuggestionIndex >= 0 && suggestions[selectedSuggestionIndex]
            ? suggestions[selectedSuggestionIndex]
            : trimmedValue,
        )
        return
      }
      case 'Tab': {
        if (e.shiftKey) {
          return
        }
        e.preventDefault()
        const trimmedValue = attendees[index].name.trim()
        if (!trimmedValue.length) {
          return
        }
        handleSelectSuggestion(
          index,
          selectedSuggestionIndex >= 0 && suggestions[selectedSuggestionIndex]
            ? suggestions[selectedSuggestionIndex]
            : trimmedValue,
        )
      }
    }
  }

  const onSubmit = (data: FormValues) => {
    alert(
      'Attendance submitted: ' + data.attendees.map((it) => it.name).join(', '),
    )
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-6">
      <div className="flex flex-col gap-4">
        {fields.map((field, index) => {
          const isFocused = index === activeInputIndex
          return (
            <div key={field._id} className="relative">
              <div className="flex items-center gap-2">
                <div className="relative w-full">
                  <Controller
                    name={`attendees.${index}.name`}
                    control={control}
                    render={({ field }) => {
                      const isDuplicate =
                        !!field.value.length &&
                        attendees.findIndex(
                          (it) =>
                            it.name.toLowerCase() === field.value.toLowerCase(),
                        ) !== index
                      const isNewName =
                        field.value.trim() !== '' &&
                        !people.some(
                          (person) =>
                            person.toLowerCase() === field.value.toLowerCase(),
                        )
                      const isMissingEmail = !!field.value.length // TODO(jh): implement once we have the real activist list
                      const isMissingPhone = !!field.value.length // TODO(jh): implement once we have the real activist list
                      return (
                        <div className="relative">
                          <Input
                            {...field}
                            ref={(el) => {
                              inputRefs.current[index] = el
                              field.ref(el)
                            }}
                            onChange={(e) => {
                              field.onChange(e)
                              handleInputChange(index, e.target.value)
                            }}
                            onKeyDown={(e) => handleKeyDown(e, index)}
                            placeholder="Enter name..."
                            className={cn(
                              'w-full transition-colors duration-300 border-2',
                              isDuplicate
                                ? 'text-red-500 border-red-500 focus:border-red-500'
                                : isNewName
                                  ? 'border-purple-500 focus:border-transparent'
                                  : '',
                            )}
                            autoComplete="off"
                            onFocus={() => setActiveInputIndex(index)}
                            onBlur={() => {
                              // If there are less than 2 empty rows, add another row.
                              if (
                                attendees.filter((it) => !it.name.length)
                                  .length < MIN_EMPTY_FIELDS
                              ) {
                                append({ name: '' }, { shouldFocus: false })
                              }
                            }}
                          />
                          <div className="right-0 top-0 bottom-0 h-full pointer-events-none absolute flex gap-2 items-center p-1.5 opacity-80">
                            {isNewName && (
                              <UserRoundPlus className="text-purple-500" />
                            )}
                            {/* TODO(jh): in a perfect world, maybe tapping these icons could open a modal to add the missing info (or a qr code to scan that prefills sthe name?) or maybe the check-in form can just automatically ask people if they are at at event, if an event is in progress, then add them to attendance automatically? */}
                            {!isNewName && isMissingEmail && (
                              <MailX className="text-orange-500" />
                            )}
                            {!isNewName && isMissingPhone && (
                              <PhoneMissed className="text-orange-500" />
                            )}
                          </div>
                        </div>
                      )
                    }}
                  />
                  {isFocused && !!suggestions.length && (
                    <ul className="absolute z-10 mt-1 w-full rounded-md border border-gray-200 bg-white shadow-lg">
                      {suggestions.map((suggestion, i) => (
                        <li
                          key={suggestion}
                          className={cn(
                            'cursor-pointer px-4 py-2 hover:bg-gray-100',
                            i === selectedSuggestionIndex
                              ? 'bg-neutral-100'
                              : '',
                          )}
                          onClick={() => {
                            handleSelectSuggestion(index, suggestion)
                          }}
                        >
                          {suggestion}
                        </li>
                      ))}
                    </ul>
                  )}
                </div>
              </div>
            </div>
          )
        })}
      </div>
      <div className="flex justify-end">
        <Button type="submit" variant="default">
          Save
        </Button>
      </div>
    </form>
  )
}
