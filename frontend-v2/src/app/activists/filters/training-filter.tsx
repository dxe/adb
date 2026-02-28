'use client'

import { Button } from '@/components/ui/button'
import { FilterChip } from './filter-chip'
import { parseIncludeExclude, buildIncludeExclude } from './filter-utils'

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
  const { include, exclude } = parseIncludeExclude(value)
  const count = include.size + exclude.size
  const hasFilter = count > 0

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
    <FilterChip
      label="Training"
      summary={hasFilter ? `${count} selected` : undefined}
      onClear={() => onChange(undefined)}
      defaultOpen={defaultOpen}
      removable={removable}
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
    </FilterChip>
  )
}
