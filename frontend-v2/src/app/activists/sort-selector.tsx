'use client'

import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { ArrowDownUp, ArrowDown, ArrowUp, ChevronDown, X } from 'lucide-react'
import { ActivistColumnName } from '@/lib/api'
import type { SortColumn } from './query-utils'
import { COLUMN_DEFINITIONS } from './column-definitions'

interface SortSelectorProps {
  /** Small label text shown above the value when active. */
  label: string
  /** Text shown on the button when no value is selected. Defaults to label. */
  inactiveLabel?: string
  value?: SortColumn
  onChange: (value: SortColumn) => void
  onClear: () => void
  /** Whether the "x" clear button is shown. */
  canClear?: boolean
  /** Columns available for sorting (subset of visible columns). */
  availableColumns: ActivistColumnName[]
}

function getColumnLabel(name: ActivistColumnName): string {
  const column = COLUMN_DEFINITIONS.find((d) => d.name === name)
  if (column === undefined) {
    throw new Error('cannot find column: ' + name)
  }

  return column.label
}

export function SortSelector({
  label,
  inactiveLabel,
  value,
  onChange,
  onClear,
  canClear = true,
  availableColumns,
}: SortSelectorProps) {
  const handleColumnChange = (column: string) => {
    onChange({
      column: column as ActivistColumnName,
      // Force ASC when selecting 'id' (DESC not allowed); otherwise preserve current direction.
      desc: column === 'id' ? false : (value?.desc ?? false),
    })
  }

  const handleDirectionChange = (desc: boolean) => {
    if (value) {
      onChange({ ...value, desc })
    }
  }

  const DirectionIcon = value?.desc ? ArrowDown : ArrowUp

  return (
    <div className="flex shrink-0 items-stretch overflow-hidden rounded-md border bg-card h-12">
      <Popover>
        <PopoverTrigger asChild>
          <button className="flex items-center gap-2 px-3 hover:bg-muted transition-colors h-full">
            <ArrowDownUp className="h-4 w-4 shrink-0 text-muted-foreground" />
            {!value ? (
              <div className="flex items-center gap-1">
                <span className="text-sm">{inactiveLabel ?? label}</span>
                <ChevronDown className="h-3 w-3 text-muted-foreground" />
              </div>
            ) : (
              <div className="flex flex-col items-start">
                <span className="text-xs text-muted-foreground">{label}</span>
                <div className="flex items-center gap-1">
                  <span className="text-sm">
                    {getColumnLabel(value.column)}
                  </span>
                  <DirectionIcon className="h-3 w-3 text-muted-foreground" />
                </div>
              </div>
            )}
          </button>
        </PopoverTrigger>
        <PopoverContent className="w-64">
          <div className="space-y-3">
            <div className="space-y-1.5">
              <Label className="text-sm font-medium">Column</Label>
              <Select
                value={value?.column ?? ''}
                onValueChange={handleColumnChange}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Choose column..." />
                </SelectTrigger>
                <SelectContent>
                  {availableColumns.map((col) => (
                    <SelectItem key={col} value={col}>
                      {getColumnLabel(col)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            {value && (
              <div className="space-y-1.5">
                <Label className="text-sm font-medium">Direction</Label>
                <div className="flex gap-1">
                  <Button
                    variant={!value.desc ? 'default' : 'outline'}
                    size="sm"
                    className="flex-1 gap-1"
                    onClick={() => handleDirectionChange(false)}
                  >
                    <ArrowUp className="h-3.5 w-3.5" />
                    Ascending
                  </Button>
                  <Button
                    variant={value.desc ? 'default' : 'outline'}
                    size="sm"
                    className="flex-1 gap-1"
                    onClick={() => handleDirectionChange(true)}
                    disabled={value.column === 'id'}
                  >
                    <ArrowDown className="h-3.5 w-3.5" />
                    Descending
                  </Button>
                </div>
              </div>
            )}
          </div>
        </PopoverContent>
      </Popover>
      {value && canClear && (
        <button
          onClick={onClear}
          className="border-l px-2 hover:bg-muted transition-colors"
          aria-label={`Clear ${label.toLowerCase()}`}
        >
          <X className="h-4 w-4" />
        </button>
      )}
    </div>
  )
}
