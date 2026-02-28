'use client'

import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { FilterChip } from './filter-chip'
import { useDraftFilter } from './filter-utils'

interface SourceFilterChipProps {
  value?: string
  onChange: (value?: string) => void
  defaultOpen?: boolean
  removable?: boolean
}

export function SourceFilterChip({
  value,
  onChange,
  defaultOpen,
  removable,
}: SourceFilterChipProps) {
  const [draft, setDraft, onOpenChange] = useDraftFilter(value, onChange)

  return (
    <FilterChip
      label="Source"
      summary={value}
      onClear={() => onChange(undefined)}
      defaultOpen={defaultOpen}
      removable={removable}
      onOpenChange={onOpenChange}
    >
      <div className="space-y-3">
        <h4 className="font-medium text-sm">Source</h4>
        <Input
          value={draft || ''}
          onChange={(e) => setDraft(e.target.value || undefined)}
          placeholder="e.g. form,petition,-application"
        />
        <p className="text-xs text-muted-foreground">
          Comma-separated. Prefix with - to exclude.
        </p>
        {draft && (
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
