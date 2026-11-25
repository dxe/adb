'use client'

import { FilterChip } from './filter-chip'
import { IncludeExcludeOption } from './include-exclude-option'
import { useDraftFilter } from '@/hooks/use-draft-filter'
import type { IncludeExcludeFilterValue } from '../filter-types'

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
  value?: IncludeExcludeFilterValue
  onChange: (value?: IncludeExcludeFilterValue) => void
  defaultOpen?: boolean
  removable?: boolean
}

function normalizeIncludeExclude(
  value?: IncludeExcludeFilterValue,
): IncludeExcludeFilterValue | undefined {
  if (!value) return undefined

  const includeSet = new Set(value.include)
  const excludeSet = new Set(value.exclude)
  for (const v of includeSet) {
    excludeSet.delete(v)
  }
  const include = Array.from(includeSet)
  const exclude = Array.from(excludeSet)

  if (include.length === 0 && exclude.length === 0) return undefined

  return { include, exclude }
}

export function TrainingFilter({
  value,
  onChange,
  defaultOpen,
  removable,
}: TrainingFilterProps) {
  const [draft, setDraft, chipProps] = useDraftFilter(value, onChange)
  const include = new Set(draft?.include ?? [])
  const exclude = new Set(draft?.exclude ?? [])

  const updateDraft = (next: IncludeExcludeFilterValue | undefined) => {
    setDraft(normalizeIncludeExclude(next))
  }

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

    updateDraft({
      include: Array.from(newInclude),
      exclude: Array.from(newExclude),
    })
  }

  // Summary is derived from the committed value, not the draft.
  const committedCount =
    (value?.include.length ?? 0) + (value?.exclude.length ?? 0)
  const summary = committedCount > 0 ? `${committedCount} selected` : undefined

  return (
    <FilterChip
      label="Training"
      summary={summary}
      onClear={() => onChange(undefined)}
      defaultOpen={defaultOpen}
      removable={removable}
      {...chipProps}
    >
      <div className="space-y-3">
        <h4 className="font-medium text-sm">Training</h4>
        <p className="text-xs text-muted-foreground">
          Click to require completed, click again to require not completed,
          click again to clear.
        </p>
        {TRAINING_COLUMNS.map(({ value: col, label }) => {
          // Unlike the other include/exclude filters, each training has its
          // own mode.
          const isExcluded = exclude.has(col)
          return (
            <IncludeExcludeOption
              key={col}
              label={label}
              selected={include.has(col) || isExcluded}
              mode={isExcluded ? 'exclude' : 'include'}
              onToggle={() => handleToggle(col)}
            />
          )
        })}
      </div>
    </FilterChip>
  )
}
