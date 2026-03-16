'use client'

import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import {
  decodeIncludeExcludeSet,
  encodeIncludeExclude,
} from '../filter-url-codecs'
import { FilterChip } from './filter-chip'
import { useDraftFilter } from '@/hooks/use-draft-filter'
import type { IncludeExcludeFilterValue } from '../filter-types'

interface SourceFilterChipProps {
  value?: IncludeExcludeFilterValue
  onChange: (value?: IncludeExcludeFilterValue) => void
  defaultOpen?: boolean
  removable?: boolean
}

function formatSourceValue(
  value?: IncludeExcludeFilterValue,
): string | undefined {
  return encodeIncludeExclude(value?.include ?? [], value?.exclude ?? [])
}

function parseSourceValue(raw?: string): IncludeExcludeFilterValue | undefined {
  const parsed = decodeIncludeExcludeSet(raw)
  if (!parsed) return undefined
  return {
    include: Array.from(parsed.include),
    exclude: Array.from(parsed.exclude),
  }
}

export function SourceFilterChip({
  value,
  onChange,
  defaultOpen,
  removable,
}: SourceFilterChipProps) {
  const committedText = formatSourceValue(value)
  const [draft, setDraft, onOpenChange] = useDraftFilter(
    committedText,
    (next) => onChange(parseSourceValue(next)),
  )

  return (
    <FilterChip
      label="Source"
      summary={committedText}
      summaryClassName="max-w-[15ch]"
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
