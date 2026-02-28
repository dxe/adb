'use client'

import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { DatePicker } from '@/components/ui/date-picker'
import { ChevronDown, X } from 'lucide-react'
import { format } from 'date-fns'

interface LastEventFilterProps {
  lastEventGte?: string // ISO date string (YYYY-MM-DD)
  lastEventLt?: string // ISO date string (YYYY-MM-DD)
  onChange: (gte?: string, lt?: string) => void
}

export function LastEventFilter({
  lastEventGte,
  lastEventLt,
  onChange,
}: LastEventFilterProps) {
  const hasFilter = lastEventGte || lastEventLt

  const formatDateRange = () => {
    const currentYear = new Date().getFullYear()
    if (lastEventGte && lastEventLt) {
      const gteDate = toUtcDate(lastEventGte)
      const ltDate = toUtcDate(lastEventLt)
      const bothCurrentYear =
        gteDate.getFullYear() === currentYear &&
        ltDate.getFullYear() === currentYear
      return bothCurrentYear
        ? `${format(gteDate, 'MMM d')} - ${format(ltDate, 'MMM d')}`
        : `${format(gteDate, 'MMM d, yyyy')} - ${format(ltDate, 'MMM d, yyyy')}`
    } else if (lastEventGte) {
      return `On or after ${format(toUtcDate(lastEventGte), 'MMM d, yyyy')}`
    } else if (lastEventLt) {
      return `Before ${format(toUtcDate(lastEventLt), 'MMM d, yyyy')}`
    } else {
      return null
    }
  }

  const handleClear = () => {
    onChange(undefined, undefined)
  }

  const toUtcDate = (dateString: string) => new Date(`${dateString}T00:00:00`)

  const toDateString = (date?: Date) =>
    date
      ? `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
      : undefined

  const handleGteChange = (date?: Date) => {
    onChange(toDateString(date), lastEventLt)
  }

  const handleLtChange = (date?: Date) => {
    onChange(lastEventGte, toDateString(date))
  }

  return (
    <div className="flex shrink-0 items-stretch rounded-md border bg-card overflow-hidden h-12">
      <Popover>
        <PopoverTrigger asChild>
          <button className="flex flex-col items-start justify-center px-3 hover:bg-muted transition-colors h-full">
            {!hasFilter ? (
              <div className="flex items-center gap-1">
                <span className="text-sm">Last event</span>
                <ChevronDown className="h-3 w-3 text-muted-foreground" />
              </div>
            ) : (
              <>
                <span className="text-xs text-muted-foreground">
                  Last event
                </span>
                <div className="flex items-center gap-1">
                  <span className="text-sm">{formatDateRange()}</span>
                  <ChevronDown className="h-3 w-3 text-muted-foreground" />
                </div>
              </>
            )}
          </button>
        </PopoverTrigger>
        <PopoverContent className="w-80">
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="last-event-after" className="text-sm font-medium">
                On or after
              </Label>
              <DatePicker
                value={lastEventGte ? toUtcDate(lastEventGte) : undefined}
                onValueChange={handleGteChange}
                placeholder="Select start date"
              />
            </div>
            <div className="space-y-2">
              <Label
                htmlFor="last-event-before"
                className="text-sm font-medium"
              >
                Before
              </Label>
              <DatePicker
                value={lastEventLt ? toUtcDate(lastEventLt) : undefined}
                onValueChange={handleLtChange}
                placeholder="Select end date"
              />
            </div>
            {hasFilter && (
              <Button
                variant="outline"
                size="sm"
                className="w-full"
                onClick={handleClear}
              >
                Clear dates
              </Button>
            )}
          </div>
        </PopoverContent>
      </Popover>
      {hasFilter && (
        <button
          onClick={handleClear}
          className="border-l px-2 hover:bg-muted transition-colors"
          aria-label="Clear filter"
        >
          <X className="h-4 w-4" />
        </button>
      )}
    </div>
  )
}
