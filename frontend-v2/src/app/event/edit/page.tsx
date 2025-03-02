'use client'

import { useRef, useState, KeyboardEvent } from 'react'
import { useForm, useFieldArray, Controller } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

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

export default function AttendancePage() {
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
  const [duplicateIndex, setDuplicateIndex] = useState<number | null>(null)

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

    // Check for duplicates.
    const firstIndex = attendees.findIndex(
      (it) => it.name.toLowerCase() === value.toLowerCase(),
    )
    if (firstIndex !== index) {
      setDuplicateIndex(index)
      inputRefs.current[index]?.focus()
      setActiveInputIndex(index)
      setTimeout(() => {
        setDuplicateIndex(null)
        // Ensure the input remains focused
      }, 300)
      return
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
        if (index < fields.length - 1) {
          inputRefs.current[index + 1]?.focus()
        }
        return
      }
    }
  }

  const onSubmit = (data: FormValues) => {
    alert(
      'Attendance submitted: ' + data.attendees.map((it) => it.name).join(', '),
    )
  }

  const isNewName = (name: string) => {
    return (
      name.trim() !== '' &&
      !people.some((person) => person.toLowerCase() === name.toLowerCase())
    )
  }

  return (
    <div className="py-4 md:py-8 px-2 md:px-4 flex justify-center">
      <div className="max-w-lg flex flex-col gap-6">
        <h1 className="text-3xl font-bold">Attendance</h1>
        <p className="text-muted-foreground">
          Enter the names of attendees below. Type to see suggestions or enter
          new names.
        </p>
        <form onSubmit={handleSubmit(onSubmit)}>
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
                        render={({ field }) => (
                          <Input
                            {...field}
                            ref={(el) => {
                              inputRefs.current[index] = el
                              field.ref(el)
                            }}
                            onChange={(e) => {
                              if (duplicateIndex) {
                                return
                              }
                              field.onChange(e)
                              handleInputChange(index, e.target.value)
                            }}
                            onKeyDown={(e) => handleKeyDown(e, index)}
                            placeholder="Enter name..."
                            className={cn(
                              'w-full transition-colors duration-300',
                              !isFocused &&
                                isNewName(field.value) &&
                                'border-blue-500 border-2',
                              duplicateIndex === index &&
                                'border-red-500 text-red-500 animate-shake border-2',
                            )}
                            autoComplete="off"
                            onFocus={() => setActiveInputIndex(index)}
                            onBlur={() => {
                              setActiveInputIndex(-1)
                              // If a suggestion is highlighted, use it.
                              if (
                                selectedSuggestionIndex >= 0 &&
                                suggestions[selectedSuggestionIndex]
                              ) {
                                handleSelectSuggestion(
                                  index,
                                  suggestions[selectedSuggestionIndex],
                                )
                              } else if (attendees[index].name.trim() !== '') {
                                // Otherwise, use the input value.
                                handleSelectSuggestion(
                                  index,
                                  attendees[index].name.trim(),
                                )
                              }
                              // If there are less than 2 more rows below us, add another row.
                              if (
                                attendees.filter((it) => !it.name.length)
                                  .length < MIN_EMPTY_FIELDS
                              ) {
                                append({ name: '' }, { shouldFocus: false })
                              }
                            }}
                          />
                        )}
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
                                if (index < fields.length - 1) {
                                  inputRefs.current[index + 1]?.focus()
                                }
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
          <div className="flex justify-end py-6">
            <Button type="submit" variant="default">
              Save
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}
