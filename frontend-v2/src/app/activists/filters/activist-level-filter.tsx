'use client'

import { useReducer } from 'react'
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
  ACTIVIST_LEVELS,
  type ActivistLevelFilterValue,
  type ActivistLevelValue,
} from '../filter-types'

interface ActivistLevelFilterProps {
  value?: ActivistLevelFilterValue
  onChange: (value?: ActivistLevelFilterValue) => void
}

type DraftState = {
  mode: 'include' | 'exclude'
  values: Set<ActivistLevelValue>
}

type Action =
  | { type: 'reset'; value?: ActivistLevelFilterValue }
  | { type: 'setMode'; mode: 'include' | 'exclude' }
  | { type: 'toggle'; level: ActivistLevelValue }
  | { type: 'clear' }

function fromValue(value?: ActivistLevelFilterValue): DraftState {
  if (!value) return { mode: 'include', values: new Set<ActivistLevelValue>() }
  return { mode: value.mode, values: new Set(value.values) }
}

function toValue(state: DraftState): ActivistLevelFilterValue | undefined {
  const values = Array.from(state.values)
  if (values.length === 0) return undefined
  return {
    mode: state.mode,
    values,
  }
}

function isSameValue(
  a?: ActivistLevelFilterValue,
  b?: ActivistLevelFilterValue,
): boolean {
  if (!a && !b) return true
  if (!a || !b) return false
  if (a.mode !== b.mode) return false
  if (a.values.length !== b.values.length) return false

  const aSorted = [...a.values].sort()
  const bSorted = [...b.values].sort()
  return aSorted.every((value, index) => value === bSorted[index])
}

function reducer(state: DraftState, action: Action): DraftState {
  switch (action.type) {
    case 'reset':
      return fromValue(action.value)
    case 'setMode':
      return { ...state, mode: action.mode }
    case 'toggle': {
      const values = new Set(state.values)
      if (values.has(action.level)) {
        values.delete(action.level)
      } else {
        values.add(action.level)
      }
      return { ...state, values }
    }
    case 'clear':
      return { ...state, values: new Set<ActivistLevelValue>() }
  }
}

export function ActivistLevelFilter({
  value,
  onChange,
}: ActivistLevelFilterProps) {
  const [draft, dispatch] = useReducer(reducer, value, fromValue)
  const hasFilter = draft.values.size > 0

  const onOpenChange = (open: boolean) => {
    if (open) {
      dispatch({ type: 'reset', value })
    } else {
      const nextValue = toValue(draft)
      if (!isSameValue(nextValue, value)) {
        onChange(nextValue)
      }
    }
  }

  const handleModeChange = (newMode: 'include' | 'exclude') => {
    if (newMode === draft.mode) return
    dispatch({ type: 'setMode', mode: newMode })
  }

  const handleToggle = (level: ActivistLevelValue) => {
    dispatch({ type: 'toggle', level })
  }

  // Summary is derived from the committed value, not the draft.
  const committed = fromValue(value)
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
            value={draft.mode}
            onValueChange={(v) => handleModeChange(v as 'include' | 'exclude')}
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
          const isSelected = draft.values.has(level)
          return (
            <button
              type="button"
              key={level}
              className="flex w-full items-center gap-2 rounded px-2 py-1 text-sm hover:bg-muted transition-colors"
              onClick={() => handleToggle(level)}
            >
              <span
                className={`flex h-4 w-4 shrink-0 items-center justify-center rounded border text-xs font-bold ${
                  isSelected
                    ? draft.mode === 'include'
                      ? 'bg-primary text-primary-foreground border-primary'
                      : 'bg-destructive text-destructive-foreground border-destructive'
                    : 'border-input'
                }`}
              >
                {isSelected ? (draft.mode === 'include' ? '+' : '-') : ''}
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
            onClick={() => dispatch({ type: 'clear' })}
          >
            Clear
          </Button>
        )}
      </div>
    </FilterChip>
  )
}
