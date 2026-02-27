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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { SlidersHorizontal, ChevronDown, X, Users } from 'lucide-react'
import { LastEventFilter } from './last-event-filter'
import { DateRangeFilter } from './date-range-filter'
import type { FilterState } from '../query-utils'

interface ActivistFiltersProps {
  filters: FilterState
  onFiltersChange: (filters: FilterState) => void
  isAdmin: boolean
  children?: React.ReactNode
}

const ACTIVIST_LEVELS = [
  'Supporter',
  'Chapter Member',
  'Organizer',
  'Non-Local',
  'Global Network Member',
] as const

const TRAINING_COLUMNS = [
  { value: 'training0', label: 'Workshop (101)' },
  { value: 'training1', label: 'Consent' },
  { value: 'training4', label: 'Training 4' },
  { value: 'training5', label: 'Training 5' },
  { value: 'training6', label: 'Training 6' },
  { value: 'consent_quiz', label: 'Consent Quiz' },
  { value: 'training_protest', label: 'Protest Training' },
  { value: 'dev_quiz', label: 'Dev Quiz' },
] as const

/** Parse "a,b,-c" into include/exclude sets. */
function parseIncludeExclude(value?: string): {
  include: Set<string>
  exclude: Set<string>
} {
  const include = new Set<string>()
  const exclude = new Set<string>()
  if (!value) return { include, exclude }
  for (const part of value.split(',')) {
    const trimmed = part.trim()
    if (!trimmed) continue
    if (trimmed.startsWith('-')) {
      exclude.add(trimmed.slice(1))
    } else {
      include.add(trimmed)
    }
  }
  return { include, exclude }
}

/** Build "a,b,-c" from include/exclude sets. */
function buildIncludeExclude(
  include: Set<string>,
  exclude: Set<string>,
): string | undefined {
  const parts = [
    ...Array.from(include),
    ...Array.from(exclude).map((v) => `-${v}`),
  ]
  return parts.length > 0 ? parts.join(',') : undefined
}

/** Parse "1..4" into parts. */
function parseIntRange(value?: string): { gte?: string; lt?: string } {
  if (!value) return {}
  const parts = value.split('..')
  if (parts.length !== 2) return {}
  return { gte: parts[0] || undefined, lt: parts[1] || undefined }
}

/** Build "1..4" from parts. */
function buildIntRange(gte?: string, lt?: string): string | undefined {
  if (!gte && !lt) return undefined
  return `${gte || ''}..${lt || ''}`
}

// --- Filter sub-components ---

function ActivistLevelFilter({
  value,
  onChange,
}: {
  value?: string
  onChange: (value?: string) => void
}) {
  const { include, exclude } = parseIncludeExclude(value)
  const hasFilter = include.size > 0 || exclude.size > 0
  // Determine current mode: "include" if any includes exist, "exclude" if any excludes exist, default "include".
  const mode: 'include' | 'exclude' =
    exclude.size > 0 && include.size === 0 ? 'exclude' : 'include'
  const selected = mode === 'include' ? include : exclude

  const handleModeChange = (newMode: 'include' | 'exclude') => {
    if (newMode === mode) return
    // When switching mode, transfer all selected items to the other set.
    if (newMode === 'include') {
      onChange(buildIncludeExclude(new Set(exclude), new Set()))
    } else {
      onChange(buildIncludeExclude(new Set(), new Set(include)))
    }
  }

  const handleToggle = (level: string) => {
    const newSelected = new Set(selected)
    if (newSelected.has(level)) {
      newSelected.delete(level)
    } else {
      newSelected.add(level)
    }
    if (mode === 'include') {
      onChange(buildIncludeExclude(newSelected, new Set()))
    } else {
      onChange(buildIncludeExclude(new Set(), newSelected))
    }
  }

  return (
    <div className="flex shrink-0 items-stretch rounded-md border bg-card overflow-hidden h-12">
      <Popover>
        <PopoverTrigger asChild>
          <button className="flex flex-col items-start justify-center px-3 hover:bg-muted transition-colors h-full">
            {!hasFilter ? (
              <div className="flex items-center gap-1">
                <span className="text-sm">Activist level</span>
                <ChevronDown className="h-3 w-3 text-muted-foreground" />
              </div>
            ) : (
              <>
                <span className="text-xs text-muted-foreground">
                  Activist level
                </span>
                <div className="flex items-center gap-1">
                  <span className="text-sm">
                    {mode === 'exclude' ? 'not ' : ''}
                    {Array.from(selected).join(', ')}
                  </span>
                  <ChevronDown className="h-3 w-3 text-muted-foreground" />
                </div>
              </>
            )}
          </button>
        </PopoverTrigger>
        <PopoverContent className="w-64">
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <h4 className="font-medium text-sm">Activist Level</h4>
              <Select
                value={mode}
                onValueChange={(v) =>
                  handleModeChange(v as 'include' | 'exclude')
                }
              >
                <SelectTrigger className="h-7 w-[100px] text-xs">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="include">Include</SelectItem>
                  <SelectItem value="exclude">Exclude</SelectItem>
                </SelectContent>
              </Select>
            </div>
            {ACTIVIST_LEVELS.map((level) => {
              const isSelected = selected.has(level)
              return (
                <button
                  key={level}
                  className="flex w-full items-center gap-2 rounded px-2 py-1 text-sm hover:bg-muted transition-colors"
                  onClick={() => handleToggle(level)}
                >
                  <span
                    className={`flex h-4 w-4 shrink-0 items-center justify-center rounded border text-xs font-bold ${
                      isSelected
                        ? mode === 'include'
                          ? 'bg-primary text-primary-foreground border-primary'
                          : 'bg-destructive text-destructive-foreground border-destructive'
                        : 'border-input'
                    }`}
                  >
                    {isSelected ? (mode === 'include' ? '+' : '-') : ''}
                  </span>
                  {level}
                </button>
              )
            })}
            {hasFilter && (
              <Button
                variant="outline"
                size="sm"
                className="w-full"
                onClick={() => onChange(undefined)}
              >
                Clear
              </Button>
            )}
          </div>
        </PopoverContent>
      </Popover>
      {hasFilter && (
        <button
          onClick={() => onChange(undefined)}
          className="border-l px-2 hover:bg-muted transition-colors"
          aria-label="Clear filter"
        >
          <X className="h-4 w-4" />
        </button>
      )}
    </div>
  )
}

function IntRangeFilterControl({
  label,
  value,
  onChange,
}: {
  label: string
  value?: string
  onChange: (value?: string) => void
}) {
  const { gte, lt } = parseIntRange(value)
  const hasFilter = !!value

  return (
    <div className="flex shrink-0 items-stretch rounded-md border bg-card overflow-hidden h-12">
      <Popover>
        <PopoverTrigger asChild>
          <button className="flex flex-col items-start justify-center px-3 hover:bg-muted transition-colors h-full">
            {!hasFilter ? (
              <div className="flex items-center gap-1">
                <span className="text-sm">{label}</span>
                <ChevronDown className="h-3 w-3 text-muted-foreground" />
              </div>
            ) : (
              <>
                <span className="text-xs text-muted-foreground">{label}</span>
                <div className="flex items-center gap-1">
                  <span className="text-sm">
                    {gte && lt
                      ? `${gte} to ${lt}`
                      : gte
                        ? `>= ${gte}`
                        : `< ${lt}`}
                  </span>
                  <ChevronDown className="h-3 w-3 text-muted-foreground" />
                </div>
              </>
            )}
          </button>
        </PopoverTrigger>
        <PopoverContent className="w-64">
          <div className="space-y-4">
            <div className="space-y-2">
              <Label className="text-sm font-medium">Min (inclusive)</Label>
              <Input
                type="number"
                value={gte || ''}
                onChange={(e) =>
                  onChange(buildIntRange(e.target.value || undefined, lt))
                }
                placeholder="No minimum"
              />
            </div>
            <div className="space-y-2">
              <Label className="text-sm font-medium">Max (exclusive)</Label>
              <Input
                type="number"
                value={lt || ''}
                onChange={(e) =>
                  onChange(buildIntRange(gte, e.target.value || undefined))
                }
                placeholder="No maximum"
              />
            </div>
            {hasFilter && (
              <Button
                variant="outline"
                size="sm"
                className="w-full"
                onClick={() => onChange(undefined)}
              >
                Clear
              </Button>
            )}
          </div>
        </PopoverContent>
      </Popover>
      {hasFilter && (
        <button
          onClick={() => onChange(undefined)}
          className="border-l px-2 hover:bg-muted transition-colors"
          aria-label="Clear filter"
        >
          <X className="h-4 w-4" />
        </button>
      )}
    </div>
  )
}

function SourceFilterControl({
  value,
  onChange,
}: {
  value?: string
  onChange: (value?: string) => void
}) {
  const hasFilter = !!value

  return (
    <div className="flex shrink-0 items-stretch rounded-md border bg-card overflow-hidden h-12">
      <Popover>
        <PopoverTrigger asChild>
          <button className="flex flex-col items-start justify-center px-3 hover:bg-muted transition-colors h-full">
            {!hasFilter ? (
              <div className="flex items-center gap-1">
                <span className="text-sm">Source</span>
                <ChevronDown className="h-3 w-3 text-muted-foreground" />
              </div>
            ) : (
              <>
                <span className="text-xs text-muted-foreground">Source</span>
                <div className="flex items-center gap-1">
                  <span className="text-sm truncate max-w-[200px]">
                    {value}
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
              <Label className="text-sm font-medium">Source patterns</Label>
              <p className="text-xs text-muted-foreground">
                Comma-separated. Prefix with - to exclude.
              </p>
              <Input
                value={value || ''}
                onChange={(e) => onChange(e.target.value || undefined)}
                placeholder="e.g. form,petition,-application"
              />
            </div>
            {hasFilter && (
              <Button
                variant="outline"
                size="sm"
                className="w-full"
                onClick={() => onChange(undefined)}
              >
                Clear
              </Button>
            )}
          </div>
        </PopoverContent>
      </Popover>
      {hasFilter && (
        <button
          onClick={() => onChange(undefined)}
          className="border-l px-2 hover:bg-muted transition-colors"
          aria-label="Clear filter"
        >
          <X className="h-4 w-4" />
        </button>
      )}
    </div>
  )
}

function TrainingFilterControl({
  value,
  onChange,
}: {
  value?: string
  onChange: (value?: string) => void
}) {
  const { include, exclude } = parseIncludeExclude(value)
  const hasFilter = include.size > 0 || exclude.size > 0

  const handleToggle = (col: string) => {
    const newInclude = new Set(include)
    const newExclude = new Set(exclude)
    if (newInclude.has(col)) {
      newInclude.delete(col)
      newExclude.add(col)
    } else if (newExclude.has(col)) {
      newExclude.delete(col)
    } else {
      newInclude.add(col)
    }
    onChange(buildIncludeExclude(newInclude, newExclude))
  }

  return (
    <div className="flex shrink-0 items-stretch rounded-md border bg-card overflow-hidden h-12">
      <Popover>
        <PopoverTrigger asChild>
          <button className="flex flex-col items-start justify-center px-3 hover:bg-muted transition-colors h-full">
            {!hasFilter ? (
              <div className="flex items-center gap-1">
                <span className="text-sm">Training</span>
                <ChevronDown className="h-3 w-3 text-muted-foreground" />
              </div>
            ) : (
              <>
                <span className="text-xs text-muted-foreground">Training</span>
                <div className="flex items-center gap-1">
                  <span className="text-sm">
                    {include.size + exclude.size} selected
                  </span>
                  <ChevronDown className="h-3 w-3 text-muted-foreground" />
                </div>
              </>
            )}
          </button>
        </PopoverTrigger>
        <PopoverContent className="w-64">
          <div className="space-y-3">
            <h4 className="font-medium text-sm">Training</h4>
            <p className="text-xs text-muted-foreground">
              Click to require completed, click again to require not completed,
              click again to clear.
            </p>
            {TRAINING_COLUMNS.map(({ value: col, label }) => {
              const isIncluded = include.has(col)
              const isExcluded = exclude.has(col)
              return (
                <button
                  key={col}
                  className="flex w-full items-center gap-2 rounded px-2 py-1 text-sm hover:bg-muted transition-colors"
                  onClick={() => handleToggle(col)}
                >
                  <span
                    className={`flex h-4 w-4 shrink-0 items-center justify-center rounded border text-xs font-bold ${
                      isIncluded
                        ? 'bg-primary text-primary-foreground border-primary'
                        : isExcluded
                          ? 'bg-destructive text-destructive-foreground border-destructive'
                          : 'border-input'
                    }`}
                  >
                    {isIncluded ? '+' : isExcluded ? '-' : ''}
                  </span>
                  {label}
                </button>
              )
            })}
            {hasFilter && (
              <Button
                variant="outline"
                size="sm"
                className="w-full"
                onClick={() => onChange(undefined)}
              >
                Clear
              </Button>
            )}
          </div>
        </PopoverContent>
      </Popover>
      {hasFilter && (
        <button
          onClick={() => onChange(undefined)}
          className="border-l px-2 hover:bg-muted transition-colors"
          aria-label="Clear filter"
        >
          <X className="h-4 w-4" />
        </button>
      )}
    </div>
  )
}

export function ActivistFilters({
  filters,
  onFiltersChange,
  isAdmin,
  children,
}: ActivistFiltersProps) {
  const prospectFilterCount = [
    filters.assignedTo,
    filters.followups,
    filters.prospect,
    filters.interestDate,
    filters.totalInteractions,
  ].filter(Boolean).length

  const moreFilterCount = [
    filters.searchAcrossChapters,
    filters.includeHidden,
  ].filter(Boolean).length

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

        {/* Event filters grouped together: last event, first event, total events */}
        <LastEventFilter
          value={filters.lastEvent}
          onChange={(value) =>
            onFiltersChange({ ...filters, lastEvent: value })
          }
        />

        <DateRangeFilter
          label="First event"
          value={filters.firstEvent}
          onChange={(value) =>
            onFiltersChange({ ...filters, firstEvent: value })
          }
        />

        <IntRangeFilterControl
          label="Total events"
          value={filters.totalEvents}
          onChange={(value) =>
            onFiltersChange({ ...filters, totalEvents: value })
          }
        />

        {/* Activist level */}
        <ActivistLevelFilter
          value={filters.activistLevel}
          onChange={(value) =>
            onFiltersChange({ ...filters, activistLevel: value })
          }
        />

        {/* Source */}
        <SourceFilterControl
          value={filters.source}
          onChange={(value) => onFiltersChange({ ...filters, source: value })}
        />

        {/* Training */}
        <TrainingFilterControl
          value={filters.training}
          onChange={(value) => onFiltersChange({ ...filters, training: value })}
        />

        {/* Prospects popover */}
        <Popover>
          <PopoverTrigger asChild>
            <Button variant="outline" size="sm" className="h-12 gap-2 shrink-0">
              <Users className="h-4 w-4" />
              Prospects
              {prospectFilterCount > 0 && (
                <span className="ml-1 flex h-5 min-w-5 items-center justify-center rounded-full bg-secondary px-1.5 text-xs font-medium text-secondary-foreground">
                  {prospectFilterCount}
                </span>
              )}
            </Button>
          </PopoverTrigger>
          <PopoverContent className="w-80">
            <div className="space-y-4">
              <h4 className="font-medium">Prospects</h4>

              {/* Assigned To */}
              <div className="space-y-2">
                <Label className="text-sm font-medium">Assigned to</Label>
                <Select
                  value={filters.assignedTo || '_none'}
                  onValueChange={(v) =>
                    onFiltersChange({
                      ...filters,
                      assignedTo: v === '_none' ? undefined : v,
                    })
                  }
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Not filtered" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="_none">Not filtered</SelectItem>
                    <SelectItem value="me">Assigned to me</SelectItem>
                    <SelectItem value="any">Any assignee</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              {/* Followups */}
              <div className="space-y-2">
                <Label className="text-sm font-medium">Follow-ups</Label>
                <Select
                  value={filters.followups || '_none'}
                  onValueChange={(v) =>
                    onFiltersChange({
                      ...filters,
                      followups: v === '_none' ? undefined : v,
                    })
                  }
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Not filtered" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="_none">Not filtered</SelectItem>
                    <SelectItem value="all">All with follow-up</SelectItem>
                    <SelectItem value="due">Due</SelectItem>
                    <SelectItem value="upcoming">Upcoming</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              {/* Prospect */}
              <div className="space-y-2">
                <Label className="text-sm font-medium">Prospect</Label>
                <Select
                  value={filters.prospect || '_none'}
                  onValueChange={(v) =>
                    onFiltersChange({
                      ...filters,
                      prospect: v === '_none' ? undefined : v,
                    })
                  }
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Not filtered" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="_none">Not filtered</SelectItem>
                    <SelectItem value="chapterMember">
                      Chapter Member Prospect
                    </SelectItem>
                    <SelectItem value="organizer">
                      Organizer Prospect
                    </SelectItem>
                  </SelectContent>
                </Select>
              </div>

              {/* Interest date */}
              <div className="space-y-2">
                <Label className="text-sm font-medium">Interest date</Label>
                <DateRangeFilter
                  label="Interest date"
                  value={filters.interestDate}
                  onChange={(value) =>
                    onFiltersChange({ ...filters, interestDate: value })
                  }
                />
              </div>

              {/* Total interactions */}
              <div className="space-y-2">
                <Label className="text-sm font-medium">
                  Total interactions
                </Label>
                <IntRangeFilterControl
                  label="Total interactions"
                  value={filters.totalInteractions}
                  onChange={(value) =>
                    onFiltersChange({
                      ...filters,
                      totalInteractions: value,
                    })
                  }
                />
              </div>
            </div>
          </PopoverContent>
        </Popover>

        {/* More filters popover */}
        <Popover>
          <PopoverTrigger asChild>
            <Button variant="outline" size="sm" className="h-12 gap-2 shrink-0">
              <SlidersHorizontal className="h-4 w-4" />
              More filters
              {moreFilterCount > 0 && (
                <span className="ml-1 flex h-5 min-w-5 items-center justify-center rounded-full bg-secondary px-1.5 text-xs font-medium text-secondary-foreground">
                  {moreFilterCount}
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

export type { FilterState } from '../query-utils'
