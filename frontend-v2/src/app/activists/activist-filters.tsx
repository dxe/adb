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
import { SlidersHorizontal } from 'lucide-react'
import { LastEventFilter } from './last-event-filter'
import type { FilterState } from './query-utils'

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
      {/* Name search */}
      <Input
        id="name-search"
        type="text"
        placeholder="Search activists by name..."
        value={filters.nameSearch}
        onChange={(e) =>
          onFiltersChange({ ...filters, nameSearch: e.target.value })
        }
      />

      {/* Filter buttons with vertical wrapping */}
      <div className="flex flex-wrap items-center gap-2">
        {/* Non-filter components */}
        {children}

        {/* Last event filter */}
        <LastEventFilter
          lastEventGte={filters.lastEventGte}
          lastEventLt={filters.lastEventLt}
          onChange={(gte, lt) =>
            onFiltersChange({ ...filters, lastEventGte: gte, lastEventLt: lt })
          }
        />

        {/* More filters dropdown */}
        <Popover>
          <PopoverTrigger asChild>
            <Button variant="outline" size="sm" className="h-12 gap-2 shrink-0">
              <SlidersHorizontal className="h-4 w-4" />
              More filters
              {(filters.searchAcrossChapters || filters.includeHidden) && (
                <span className="ml-1 flex h-5 min-w-5 items-center justify-center rounded-full bg-secondary px-1.5 text-xs font-medium text-secondary-foreground">
                  {
                    [
                      filters.searchAcrossChapters,
                      filters.includeHidden,
                    ].filter(Boolean).length
                  }
                </span>
              )}
            </Button>
          </PopoverTrigger>
          <PopoverContent className="w-80">
            <div className="space-y-4">
              <h4 className="font-medium">More Filters</h4>

              {/* Search across chapters - only for admins */}
              {isAdmin && (
                <div className="flex items-center gap-2">
                  <Checkbox
                    id="search-across-chapters"
                    checked={filters.searchAcrossChapters}
                    onCheckedChange={(checked) =>
                      onFiltersChange({
                        ...filters,
                        searchAcrossChapters: Boolean(checked),
                      })
                    }
                  />
                  <Label
                    htmlFor="search-across-chapters"
                    className="cursor-pointer text-sm font-normal"
                  >
                    Search across chapters
                  </Label>
                </div>
              )}

              {/* Include hidden activists */}
              <div className="flex items-center gap-2">
                <Checkbox
                  id="include-hidden"
                  checked={filters.includeHidden}
                  onCheckedChange={(checked) =>
                    onFiltersChange({
                      ...filters,
                      includeHidden: Boolean(checked),
                    })
                  }
                />
                <Label
                  htmlFor="include-hidden"
                  className="cursor-pointer text-sm font-normal"
                >
                  Include hidden activists
                </Label>
              </div>
            </div>
          </PopoverContent>
        </Popover>
      </div>
    </div>
  )
}

export type { FilterState } from './query-utils'
