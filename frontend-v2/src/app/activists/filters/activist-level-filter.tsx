'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { FilterChip } from './filter-chip'
import {
  useDraftFilter,
  parseActivistLevelValue,
  buildActivistLevelValue,
} from './filter-utils'

const ACTIVIST_LEVELS = [
  'Supporter',
  'Chapter Member',
  'Organizer',
  'Non-Local',
  'Global Network Member',
] as const

interface ActivistLevelFilterProps {
  value?: string
  onChange: (value?: string) => void
}

export function ActivistLevelFilter({
  value,
  onChange,
}: ActivistLevelFilterProps) {
  const [draft, setDraft, onOpenChange] = useDraftFilter(value, onChange)
  const [emptyMode, setEmptyMode] = useState<'include' | 'exclude'>('include')
  const parsed = parseActivistLevelValue(draft)
  const hasFilter = parsed.values.size > 0
  const mode: 'include' | 'exclude' = hasFilter ? parsed.mode : emptyMode
  const selected = parsed.values

  const handleModeChange = (newMode: 'include' | 'exclude') => {
    setEmptyMode(newMode)
    if (newMode === mode) return
    if (!hasFilter) return
    setDraft(buildActivistLevelValue(newMode, new Set(selected)))
  }

  const handleToggle = (level: string) => {
    const newSelected = new Set(selected)
    if (newSelected.has(level)) {
      newSelected.delete(level)
    } else {
      newSelected.add(level)
    }
    setDraft(buildActivistLevelValue(mode, newSelected))
  }

  // Summary is derived from the committed value, not the draft.
  const committed = parseActivistLevelValue(value)
  const summary =
    committed.values.size > 0
      ? `${committed.mode === 'exclude' ? 'not ' : ''}${Array.from(committed.values).join(', ')}`
      : undefined

  return (
    <FilterChip
      label="Activist level"
      summary={summary}
      onClear={() => onChange(undefined)}
      onOpenChange={onOpenChange}
    >
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
            onClick={() => setDraft(undefined)}
          >
            Clear
          </Button>
        )}
      </div>
    </FilterChip>
  )
}
