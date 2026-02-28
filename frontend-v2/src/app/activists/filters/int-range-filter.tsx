'use client'

import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { FilterChip } from './filter-chip'
import { parseIntRange, buildIntRange } from './filter-utils'

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
  const { gte, lt } = parseIntRange(value)
  const hasFilter = !!value

  return (
    <FilterChip
      label={label}
      summary={hasFilter ? formatIntRange(gte, lt) : undefined}
      onClear={() => onChange(undefined)}
      defaultOpen={defaultOpen}
      removable={removable}
    >
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
    </FilterChip>
  )
}
