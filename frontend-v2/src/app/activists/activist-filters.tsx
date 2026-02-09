'use client'

import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Checkbox } from '@/components/ui/checkbox'
import { Button } from '@/components/ui/button'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { DatePicker } from '@/components/ui/date-picker'
import { ChevronDown, X } from 'lucide-react'
import { format } from 'date-fns'

interface FilterState {
  showAllChapters: boolean
  nameSearch: string
  lastEventLt?: string // ISO date string
  lastEventGte?: string // ISO date string
  // Future filters to be added:
  // includeHidden?: boolean
}

interface ActivistFiltersProps {
  filters: FilterState
  onFiltersChange: (filters: FilterState) => void
  isAdmin: boolean
  children?: React.ReactNode
}

export function ActivistFilters({
  filters,
  onFiltersChange,
  isAdmin,
  children,
}: ActivistFiltersProps) {
  return (
    <div className="flex flex-col gap-4">
      {/* Name search - always visible */}
      <Input
        id="name-search"
        type="text"
        placeholder="Search activists by name..."
        value={filters.nameSearch}
        onChange={(e) =>
          onFiltersChange({ ...filters, nameSearch: e.target.value })
        }
      />

      {/* Horizontally scrolling filter buttons */}
      <div className="flex items-center gap-2 overflow-x-auto pb-2">
        {/* Column selector or other options passed as children */}
        {children}

        {/* Chapter filter - only for admins */}
        {isAdmin && (
          <div className="flex shrink-0 items-center gap-2 rounded-md border bg-card px-3 h-12">
            <Checkbox
              id="show-all-chapters"
              checked={filters.showAllChapters}
              onCheckedChange={(checked) =>
                onFiltersChange({
                  ...filters,
                  showAllChapters: Boolean(checked),
                })
              }
            />
            <Label
              htmlFor="show-all-chapters"
              className="cursor-pointer text-sm font-normal"
            >
              Show all chapters
            </Label>
          </div>
        )}

        {/* Last Event filter chip with dropdown */}
        <div className="flex shrink-0 items-stretch rounded-md border bg-card overflow-hidden h-12">
          <Popover>
            <PopoverTrigger asChild>
              <button className="flex flex-col items-start justify-center px-3 hover:bg-muted transition-colors h-full">
                {!filters.lastEventGte && !filters.lastEventLt ? (
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
                      <span className="text-sm">
                        {(() => {
                          const currentYear = new Date().getFullYear()
                          if (filters.lastEventGte && filters.lastEventLt) {
                            const gteDate = new Date(
                              filters.lastEventGte + 'T00:00:00',
                            )
                            const ltDate = new Date(
                              filters.lastEventLt + 'T00:00:00',
                            )
                            const bothCurrentYear =
                              gteDate.getFullYear() === currentYear &&
                              ltDate.getFullYear() === currentYear
                            return bothCurrentYear
                              ? `${format(gteDate, 'MMM d')} - ${format(ltDate, 'MMM d')}`
                              : `${format(gteDate, 'MMM d, yyyy')} - ${format(ltDate, 'MMM d, yyyy')}`
                          } else if (filters.lastEventGte) {
                            return `On or after ${format(new Date(filters.lastEventGte + 'T00:00:00'), 'MMM d, yyyy')}`
                          } else {
                            return `Before ${format(new Date(filters.lastEventLt + 'T00:00:00'), 'MMM d, yyyy')}`
                          }
                        })()}
                      </span>
                      <ChevronDown className="h-3 w-3 text-muted-foreground" />
                    </div>
                  </>
                )}
              </button>
            </PopoverTrigger>
            <PopoverContent className="w-80">
              <div className="space-y-4">
                <div className="space-y-2">
                  <Label
                    htmlFor="last-event-after"
                    className="text-sm font-medium"
                  >
                    On or after
                  </Label>
                  <DatePicker
                    value={
                      filters.lastEventGte
                        ? new Date(filters.lastEventGte + 'T00:00:00')
                        : undefined
                    }
                    onValueChange={(date) => {
                      onFiltersChange({
                        ...filters,
                        lastEventGte: date
                          ? `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
                          : undefined,
                      })
                    }}
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
                    value={
                      filters.lastEventLt
                        ? new Date(filters.lastEventLt + 'T00:00:00')
                        : undefined
                    }
                    onValueChange={(date) => {
                      onFiltersChange({
                        ...filters,
                        lastEventLt: date
                          ? `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
                          : undefined,
                      })
                    }}
                    placeholder="Select end date"
                  />
                </div>
                {(filters.lastEventGte || filters.lastEventLt) && (
                  <Button
                    variant="outline"
                    size="sm"
                    className="w-full"
                    onClick={() =>
                      onFiltersChange({
                        ...filters,
                        lastEventGte: undefined,
                        lastEventLt: undefined,
                      })
                    }
                  >
                    Clear dates
                  </Button>
                )}
              </div>
            </PopoverContent>
          </Popover>
          {(filters.lastEventGte || filters.lastEventLt) && (
            <button
              onClick={() =>
                onFiltersChange({
                  ...filters,
                  lastEventGte: undefined,
                  lastEventLt: undefined,
                })
              }
              className="border-l px-2 hover:bg-muted transition-colors"
              aria-label="Clear filter"
            >
              <X className="h-4 w-4" />
            </button>
          )}
        </div>
      </div>
    </div>
  )
}

export type { FilterState }
