'use client'

import { Button } from '@/components/ui/button'
import { FilterChip } from './filter-chip'
import {
  useDraftFilter,
  parseIncludeExclude,
  buildIncludeExclude,
} from './filter-utils'

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

interface TrainingFilterProps {
  value?: string
  onChange: (value?: string) => void
  defaultOpen?: boolean
  removable?: boolean
}

export function TrainingFilter({
  value,
  onChange,
  defaultOpen,
  removable,
}: TrainingFilterProps) {
  const [draft, setDraft, onOpenChange] = useDraftFilter(value, onChange)
  const { include, exclude } = parseIncludeExclude(draft)
  const draftCount = include.size + exclude.size
  const hasDraft = draftCount > 0

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
    setDraft(buildIncludeExclude(newInclude, newExclude))
  }

  // Summary is derived from the committed value, not the draft.
  const committed = parseIncludeExclude(value)
  const committedCount = committed.include.size + committed.exclude.size
  const summary = committedCount > 0 ? `${committedCount} selected` : undefined

  return (
    <FilterChip
      label="Training"
      summary={summary}
      onClear={() => onChange(undefined)}
      defaultOpen={defaultOpen}
      removable={removable}
      onOpenChange={onOpenChange}
    >
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
        {hasDraft && (
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
