'use client'

import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { FilterChip } from './filter-chip'
import { useDraftFilter } from '@/hooks/use-draft-filter'
import type { IntRangeFilterValue } from '../filter-types'
import { parseOptionalSafeInteger } from '@/lib/number-utils'

interface IntRangeFilterProps {
  label: string
  value?: IntRangeFilterValue
  onChange: (value?: IntRangeFilterValue) => void
  defaultOpen?: boolean
  removable?: boolean
}

function normalizeIntRange(
  value?: IntRangeFilterValue,
): IntRangeFilterValue | undefined {
  if (!value) return undefined
  const gte = value.gte
  const lt = value.lt
  if (gte === undefined && lt === undefined) return undefined
  return { gte, lt }
}

function formatIntRange(gte?: number, lt?: number): string | undefined {
  if (gte !== undefined && lt !== undefined) return `[${gte} to ${lt})`
  if (gte !== undefined) return `\u2265 ${gte}`
  if (lt !== undefined) return `< ${lt}`
  return undefined
}

export function IntRangeFilter({
  label,
  value,
  onChange,
  defaultOpen,
  removable,
}: IntRangeFilterProps) {
  const [draft, setDraft, onOpenChange] = useDraftFilter(value, onChange)
  const gte = draft?.gte
  const lt = draft?.lt
  const hasDraft = !!draft

  // Summary is derived from the committed value, not the draft.
  const summary = formatIntRange(value?.gte, value?.lt)

  const updateDraft = (next: IntRangeFilterValue | undefined) => {
    setDraft(normalizeIntRange(next))
  }

  return (
    <FilterChip
      label={label}
      summary={summary}
      onClear={() => onChange(undefined)}
      defaultOpen={defaultOpen}
      removable={removable}
      onOpenChange={onOpenChange}
    >
      <div className="space-y-4">
        <div className="space-y-2">
          <Label className="text-sm font-medium">Min (inclusive)</Label>
          <Input
            type="number"
            step={1}
            inputMode="numeric"
            value={gte ?? ''}
            onChange={(e) =>
              updateDraft({
                gte: parseOptionalSafeInteger(e.target.value),
                lt,
              })
            }
            placeholder="No minimum"
          />
        </div>
        <div className="space-y-2">
          <Label className="text-sm font-medium">Max (exclusive)</Label>
          <Input
            type="number"
            step={1}
            inputMode="numeric"
            value={lt ?? ''}
            onChange={(e) =>
              updateDraft({
                gte,
                lt: parseOptionalSafeInteger(e.target.value),
              })
            }
            placeholder="No maximum"
          />
        </div>
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
