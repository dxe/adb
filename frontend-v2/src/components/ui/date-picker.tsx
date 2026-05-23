'use client'

import {
  Button,
  Calendar,
  CalendarCell,
  CalendarGrid,
  CalendarGridBody,
  CalendarGridHeader,
  CalendarHeaderCell,
  DateInput,
  DatePicker as AriaDatePicker,
  DateSegment,
  Dialog,
  Group,
  Heading,
  Popover,
} from 'react-aria-components'
import { CalendarDate } from '@internationalized/date'
import {
  CalendarIcon,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
} from 'lucide-react'
import { cn } from '@/lib/utils'

export interface DatePickerProps {
  value?: Date
  onValueChange?: (date: Date | undefined) => void
  placeholder?: string
  className?: string
  disabled?: boolean
}

function toCalendarDate(date: Date): CalendarDate {
  return new CalendarDate(
    date.getFullYear(),
    date.getMonth() + 1,
    date.getDate(),
  )
}

function toJSDate(date: CalendarDate): Date {
  return new Date(date.year, date.month - 1, date.day)
}

export function DatePicker({
  value,
  onValueChange,
  placeholder = 'Date',
  className,
  disabled,
}: DatePickerProps) {
  return (
    <AriaDatePicker
      value={value ? toCalendarDate(value) : null}
      onChange={(d) => onValueChange?.(d ? toJSDate(d) : undefined)}
      isDisabled={disabled}
      aria-label={placeholder}
      className={cn('w-full', className)}
    >
      <Group className="flex h-9 w-full items-center rounded-md border border-input bg-transparent pl-2.5 pr-3 text-sm transition-colors hover:border-gray-400 focus-within:border-primary focus-within:ring-1 focus-within:ring-ring data-[disabled]:cursor-not-allowed data-[disabled]:opacity-50">
        <span aria-hidden className="mr-2 shrink-0 text-muted-foreground">
          <CalendarIcon className="h-4 w-4" />
        </span>
        <DateInput className="flex flex-1 items-center tabular-nums">
          {(segment) => (
            <DateSegment
              segment={segment}
              className={({ isFocused, isPlaceholder }) =>
                cn(
                  'rounded outline-none',
                  segment.type === 'literal'
                    ? 'text-muted-foreground'
                    : 'px-0.5',
                  isFocused && 'bg-primary text-primary-foreground',
                  isPlaceholder && !isFocused && 'text-muted-foreground',
                )
              }
            />
          )}
        </DateInput>
        <Button
          aria-label="Open calendar"
          className="-mr-1 ml-1 shrink-0 cursor-pointer text-muted-foreground outline-none hover:text-foreground"
        >
          <ChevronDown aria-hidden className="h-4 w-4" />
        </Button>
      </Group>

      <Popover
        placement="bottom start"
        className="z-50 rounded-md border bg-popover text-popover-foreground shadow-md outline-none data-[entering]:animate-in data-[exiting]:animate-out data-[entering]:fade-in-0 data-[exiting]:fade-out-0 data-[entering]:zoom-in-95 data-[exiting]:zoom-out-95"
      >
        <Dialog className="outline-none">
          <Calendar aria-label="Calendar" className="p-3">
            <header className="relative mb-2 flex items-center justify-center">
              <Button
                slot="previous"
                aria-label="Go to previous month"
                className="absolute left-0 inline-flex h-7 w-7 items-center justify-center rounded-md hover:bg-accent disabled:opacity-50"
              >
                <ChevronLeft className="h-4 w-4" />
              </Button>
              <Heading className="text-sm font-medium" />
              <Button
                slot="next"
                aria-label="Go to next month"
                className="absolute right-0 inline-flex h-7 w-7 items-center justify-center rounded-md hover:bg-accent disabled:opacity-50"
              >
                <ChevronRight className="h-4 w-4" />
              </Button>
            </header>

            <CalendarGrid className="w-full border-collapse">
              <CalendarGridHeader>
                {(day) => (
                  <CalendarHeaderCell className="w-8 pb-1 text-center text-[0.8rem] font-normal text-muted-foreground">
                    {day}
                  </CalendarHeaderCell>
                )}
              </CalendarGridHeader>
              <CalendarGridBody>
                {(date) => (
                  <CalendarCell
                    date={date}
                    className={({
                      isSelected,
                      isToday,
                      isDisabled,
                      isFocusVisible,
                      isOutsideMonth,
                    }) =>
                      cn(
                        'flex h-8 w-8 cursor-pointer items-center justify-center rounded-md text-sm outline-none',
                        'hover:bg-accent hover:text-accent-foreground',
                        isSelected &&
                          'bg-primary text-primary-foreground hover:bg-primary hover:text-primary-foreground',
                        isToday &&
                          !isSelected &&
                          'bg-accent text-accent-foreground',
                        (isDisabled || isOutsideMonth) &&
                          'pointer-events-none opacity-50',
                        isFocusVisible && 'ring-2 ring-ring ring-offset-1',
                      )
                    }
                  />
                )}
              </CalendarGridBody>
            </CalendarGrid>
          </Calendar>
        </Dialog>
      </Popover>
    </AriaDatePicker>
  )
}
