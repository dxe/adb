'use client'

import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { FilterChip } from './filter-chip'

interface SelectFilterChipProps {
  label: string
  value?: string
  onChange: (value?: string) => void
  options: { value: string; label: string }[]
  defaultOpen?: boolean
  removable?: boolean
}

export function SelectFilterChip({
  label,
  value,
  onChange,
  options,
  defaultOpen,
  removable,
}: SelectFilterChipProps) {
  const selectedOption = options.find((o) => o.value === value)

  return (
    <FilterChip
      label={label}
      summary={selectedOption?.label}
      onClear={() => onChange(undefined)}
      defaultOpen={defaultOpen}
      removable={removable}
    >
      <div className="space-y-3">
        <h4 className="font-medium text-sm">{label}</h4>
        <Select
          value={value || '_none'}
          onValueChange={(v) => onChange(v === '_none' ? undefined : v)}
        >
          <SelectTrigger>
            <SelectValue placeholder="" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="_none">&nbsp;</SelectItem>
            {options.map((opt) => (
              <SelectItem key={opt.value} value={opt.value}>
                {opt.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        {value && (
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
