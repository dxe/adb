'use client'

import { useState, useCallback } from 'react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { Plus, X } from 'lucide-react'
import { ActivistLevelFilter } from './activist-level-filter'
import { LastEventFilter } from './last-event-filter'
import { DateRangeFilter } from './date-range-filter'
import { IntRangeFilter } from './int-range-filter'
import { TrainingFilter } from './training-filter'
import { SelectFilterChip } from './select-filter-chip'
import { SourceFilterChip } from './source-filter-chip'
import type { FilterState } from '../query-utils'

interface ActivistFiltersProps {
  filters: FilterState
  onFiltersChange: (filters: FilterState) => void
  isAdmin: boolean
  children?: React.ReactNode
}

// Keys for optional (non-pinned) filters that appear as chips when active.
type OptionalFilterKey =
  | 'firstEvent'
  | 'totalEvents'
  | 'training'
  | 'source'
  | 'interestDate'
  | 'totalInteractions'
  | 'assignedTo'
  | 'followups'
  | 'prospect'

const OPTIONAL_FILTERS: { key: OptionalFilterKey; label: string }[] = [
  { key: 'firstEvent', label: 'First event' },
  { key: 'totalEvents', label: 'Total events' },
  { key: 'training', label: 'Training' },
  { key: 'source', label: 'Source' },
  { key: 'interestDate', label: 'Interest date' },
  { key: 'totalInteractions', label: 'Total interactions' },
  { key: 'assignedTo', label: 'Assigned to' },
  { key: 'followups', label: 'Follow-ups' },
  { key: 'prospect', label: 'Prospect' },
]

/** Helper to remove a key from a Set (returns new Set). */
function without(set: Set<string>, key: string): Set<string> {
  const next = new Set(set)
  next.delete(key)
  return next
}

export function ActivistFilters({
  filters,
  onFiltersChange,
  isAdmin,
  children,
}: ActivistFiltersProps) {
  // Tracks optional filters added from the menu that may not yet have values.
  const [visibleFilters, setVisibleFilters] = useState<Set<string>>(new Set())
  const [addFilterOpen, setAddFilterOpen] = useState(false)

  const isFilterVisible = useCallback(
    (key: OptionalFilterKey) =>
      filters[key] !== undefined || visibleFilters.has(key),
    [filters, visibleFilters],
  )

  const showFilter = useCallback((key: string) => {
    // Close the popover first, then set filters on the next frame. The popover has a CSS exit animation, so it stays
    // mounted while it fades out. If the filter chips shift layout in the same render, the popover follows its trigger,
    // causing a flicker.
    setAddFilterOpen(false)
    requestAnimationFrame(() => {
      setVisibleFilters((prev) => new Set(prev).add(key))
    })
  }, [])

  /** onChange handler for optional filters — clears visibility when value is removed. */
  const optionalOnChange = useCallback(
    (key: OptionalFilterKey, value: string | undefined) => {
      onFiltersChange({ ...filters, [key]: value })
      if (value === undefined) {
        setVisibleFilters((prev) => without(prev, key))
      }
    },
    [filters, onFiltersChange],
  )

  // Filters available in the "+ Add filter" menu (not already visible).
  const availableFilters = OPTIONAL_FILTERS.filter(
    (f) => !isFilterVisible(f.key),
  )
  const hasBooleanOptions =
    (isAdmin && !filters.searchAcrossChapters) || !filters.includeHidden

  return (
    <div className="flex flex-col gap-4">
      {/* Name search */}
      <Input
        type="text"
        placeholder="Search activists by name..."
        value={filters.nameSearch}
        onChange={(e) =>
          onFiltersChange({ ...filters, nameSearch: e.target.value })
        }
      />

      {/* Chip bar */}
      <div className="flex flex-wrap items-center gap-2">
        {/* Non-filter children (columns, sort selectors) */}
        {children}

        {/* Always-visible filters */}
        <ActivistLevelFilter
          value={filters.activistLevel}
          onChange={(value) =>
            onFiltersChange({ ...filters, activistLevel: value })
          }
        />
        <LastEventFilter
          value={filters.lastEvent}
          onChange={(value) =>
            onFiltersChange({ ...filters, lastEvent: value })
          }
        />

        {/* Optional filter chips — only rendered when active or just added */}
        {isFilterVisible('firstEvent') && (
          <DateRangeFilter
            label="First event"
            value={filters.firstEvent}
            onChange={(v) => optionalOnChange('firstEvent', v)}
            removable
          />
        )}

        {isFilterVisible('totalEvents') && (
          <IntRangeFilter
            label="Total events"
            value={filters.totalEvents}
            onChange={(v) => optionalOnChange('totalEvents', v)}
            removable
          />
        )}

        {isFilterVisible('training') && (
          <TrainingFilter
            value={filters.training}
            onChange={(v) => optionalOnChange('training', v)}
            removable
          />
        )}

        {isFilterVisible('source') && (
          <SourceFilterChip
            value={filters.source}
            onChange={(v) => optionalOnChange('source', v)}
            removable
          />
        )}

        {isFilterVisible('interestDate') && (
          <DateRangeFilter
            label="Interest date"
            value={filters.interestDate}
            onChange={(v) => optionalOnChange('interestDate', v)}
            removable
          />
        )}

        {isFilterVisible('totalInteractions') && (
          <IntRangeFilter
            label="Total interactions"
            value={filters.totalInteractions}
            onChange={(v) => optionalOnChange('totalInteractions', v)}
            removable
          />
        )}

        {isFilterVisible('assignedTo') && (
          <SelectFilterChip
            label="Assigned to"
            value={filters.assignedTo}
            onChange={(v) => optionalOnChange('assignedTo', v)}
            options={[
              { value: 'me', label: 'Assigned to me' },
              { value: 'any', label: 'Any assignee' },
            ]}
            removable
          />
        )}

        {isFilterVisible('followups') && (
          <SelectFilterChip
            label="Follow-ups"
            value={filters.followups}
            onChange={(v) => optionalOnChange('followups', v)}
            options={[
              { value: 'all', label: 'All with follow-up' },
              { value: 'due', label: 'Due' },
              { value: 'upcoming', label: 'Upcoming' },
            ]}
            removable
          />
        )}

        {isFilterVisible('prospect') && (
          <SelectFilterChip
            label="Prospect"
            value={filters.prospect}
            onChange={(v) => optionalOnChange('prospect', v)}
            options={[
              { value: 'chapterMember', label: 'Chapter Member' },
              { value: 'organizer', label: 'Organizer' },
            ]}
            removable
          />
        )}

        {/* Boolean filter chips — shown in bar when active, × to deselect */}
        {filters.searchAcrossChapters && (
          <div className="flex shrink-0 items-stretch rounded-md border bg-card overflow-hidden h-12">
            <span className="flex items-center px-3 text-sm">
              Search across chapters
            </span>
            <button
              onClick={() =>
                onFiltersChange({ ...filters, searchAcrossChapters: false })
              }
              className="border-l px-2 hover:bg-muted transition-colors"
              aria-label="Disable search across chapters"
            >
              <X className="h-4 w-4" />
            </button>
          </div>
        )}

        {filters.includeHidden && (
          <div className="flex shrink-0 items-stretch rounded-md border bg-card overflow-hidden h-12">
            <span className="flex items-center px-3 text-sm">
              Include hidden
            </span>
            <button
              onClick={() =>
                onFiltersChange({ ...filters, includeHidden: false })
              }
              className="border-l px-2 hover:bg-muted transition-colors"
              aria-label="Disable include hidden"
            >
              <X className="h-4 w-4" />
            </button>
          </div>
        )}

        {/* + Add filter menu */}
        {(availableFilters.length > 0 || hasBooleanOptions) && (
          <Popover open={addFilterOpen} onOpenChange={setAddFilterOpen}>
            <PopoverTrigger asChild>
              <Button
                variant="outline"
                size="sm"
                className="h-12 gap-1 shrink-0"
              >
                <Plus className="h-4 w-4" />
                Add filter
              </Button>
            </PopoverTrigger>
            <PopoverContent className="w-52 p-2">
              <div className="flex flex-col">
                {availableFilters.map((f) => (
                  <button
                    key={f.key}
                    className="flex w-full items-center rounded px-2 py-1.5 text-sm hover:bg-muted transition-colors text-left"
                    onClick={() => showFilter(f.key)}
                  >
                    {f.label}
                  </button>
                ))}
                {hasBooleanOptions && availableFilters.length > 0 && (
                  <div className="border-t my-1" />
                )}
                {isAdmin && !filters.searchAcrossChapters && (
                  <button
                    className="flex w-full items-center rounded px-2 py-1.5 text-sm hover:bg-muted transition-colors text-left"
                    onClick={() => {
                      // Close the popover first, then set filters on the next frame. The popover has a CSS exit
                      // animation, so it stays mounted while it fades out. If the filter chips shift layout in the same
                      // render, the popover follows its trigger, causing a flicker.
                      setAddFilterOpen(false)
                      requestAnimationFrame(() => {
                        onFiltersChange({
                          ...filters,
                          searchAcrossChapters: true,
                        })
                      })
                    }}
                  >
                    Search across chapters
                  </button>
                )}
                {!filters.includeHidden && (
                  <button
                    className="flex w-full items-center rounded px-2 py-1.5 text-sm hover:bg-muted transition-colors text-left"
                    onClick={() => {
                      // Close the popover first, then set filters on the next frame. The popover has a CSS exit
                      // animation, so it stays mounted while it fades out. If the filter chips shift layout in the same
                      // render, the popover follows its trigger, causing a flicker.
                      setAddFilterOpen(false)
                      requestAnimationFrame(() => {
                        onFiltersChange({ ...filters, includeHidden: true })
                      })
                    }}
                  >
                    Include hidden
                  </button>
                )}
              </div>
            </PopoverContent>
          </Popover>
        )}
      </div>
    </div>
  )
}

export type { FilterState } from '../query-utils'
