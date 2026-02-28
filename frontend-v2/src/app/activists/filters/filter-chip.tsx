import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { ChevronDown, X } from 'lucide-react'

interface FilterChipProps {
  label: string
  /** When set, shows two-line layout (label small, summary normal) with × button. */
  summary?: string
  onClear: () => void
  /** Show × even without a summary (for optional/removable filter chips). */
  removable?: boolean
  children: React.ReactNode
  defaultOpen?: boolean
  popoverClassName?: string
}

export function FilterChip({
  label,
  summary,
  onClear,
  removable,
  children,
  defaultOpen,
  popoverClassName,
}: FilterChipProps) {
  const showClear = !!summary || removable

  return (
    <div className="flex shrink-0 items-stretch rounded-md border bg-card overflow-hidden h-12">
      <Popover defaultOpen={defaultOpen}>
        <PopoverTrigger asChild>
          <button className="flex flex-col items-start justify-center px-3 hover:bg-muted transition-colors h-full">
            {!summary ? (
              <div className="flex items-center gap-1">
                <span className="text-sm">{label}</span>
                <ChevronDown className="h-3 w-3 text-muted-foreground" />
              </div>
            ) : (
              <>
                <span className="text-xs text-muted-foreground">{label}</span>
                <div className="flex items-center gap-1">
                  <span className="text-sm">{summary}</span>
                  <ChevronDown className="h-3 w-3 text-muted-foreground" />
                </div>
              </>
            )}
          </button>
        </PopoverTrigger>
        <PopoverContent className={popoverClassName ?? 'w-64'}>
          {children}
        </PopoverContent>
      </Popover>
      {showClear && (
        <button
          onClick={onClear}
          className="border-l px-2 hover:bg-muted transition-colors"
          aria-label="Clear filter"
        >
          <X className="h-4 w-4" />
        </button>
      )}
    </div>
  )
}
