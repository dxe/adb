'use client'

import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { FilterChip } from './filter-chip'
import { useDraftFilter, parseIntRange, buildIntRange } from './filter-utils'

interface IntRangeFilterProps {
  label: string
  value?: string
  onChange: (value?: string) => void
  defaultOpen?: boolean
  removable?: boolean
}

function formatIntRange(gte?: string, lt?: string): string | undefined {
  if (gte && lt) return `[${gte} to ${lt})`
  if (gte) return `\u2265 ${gte}`
  if (lt) return `< ${lt}`
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
  const { gte, lt } = parseIntRange(draft)
  const hasDraft = !!draft

  // Summary is derived from the committed value, not the draft.
  const committed = parseIntRange(value)
  const summary = value ? formatIntRange(committed.gte, committed.lt) : undefined

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
            value={gte || ''}
            onChange={(e) =>
              setDraft(buildIntRange(e.target.value || undefined, lt))
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
              setDraft(buildIntRange(gte, e.target.value || undefined))
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
