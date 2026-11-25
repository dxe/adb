import { useState } from 'react'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { ChevronDown, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

interface FilterChipProps {
  label: string
  /** When set, shows two-line layout (label small, summary normal) with × button. */
  summary?: string
  summaryClassName?: string
  onClear: () => void
  /** Show × even without a summary (for optional/removable filter chips). */
  removable?: boolean
  children: React.ReactNode
  defaultOpen?: boolean
  popoverClassName?: string
  /** Called when the popover opens, so drafts can sync from the committed
   *  value. Closing never commits; only onApply does. */
  onOpenChange?: (open: boolean) => void
  /** Commits the draft. When set, the popover shows an Apply button and
   *  Enter in an input applies. Closing otherwise discards the draft. */
  onApply?: () => void
}

export function FilterChip({
  label,
  summary,
  summaryClassName,
  onClear,
  removable,
  children,
  defaultOpen,
  popoverClassName,
  onOpenChange,
  onApply,
}: FilterChipProps) {
  const showClear = !!summary || removable
  const [open, setOpen] = useState(defaultOpen ?? false)
  const handleOpenChange = (newOpen: boolean) => {
    setOpen(newOpen)
    if (newOpen) onOpenChange?.(true)
  }

  const handleClear = () => {
    setOpen(false)
    onClear()
  }

  const handleApply = () => {
    onApply?.()
    setOpen(false)
  }

  return (
    <div className="flex shrink-0 items-stretch rounded-md border bg-card overflow-hidden h-12">
      <Popover open={open} onOpenChange={handleOpenChange}>
        <PopoverTrigger asChild>
          <button
            type="button"
            className="flex h-full min-w-0 flex-col items-start justify-center px-3 transition-colors hover:bg-muted"
          >
            {!summary ? (
              <div className="flex items-center gap-1">
                <span className="text-sm">{label}</span>
                <ChevronDown className="h-3 w-3 text-muted-foreground" />
              </div>
            ) : (
              <>
                <span className="text-xs text-muted-foreground">{label}</span>
                <div className="flex min-w-0 items-center gap-1">
                  {/* Long summaries are cut off with an ellipsis rather than
                      stretching the chip; the title shows the full text. */}
                  <span
                    className={cn(
                      'block max-w-72 truncate text-sm',
                      summaryClassName,
                    )}
                    title={summary}
                  >
                    {summary}
                  </span>
                  <ChevronDown className="h-3 w-3 text-muted-foreground" />
                </div>
              </>
            )}
          </button>
        </PopoverTrigger>
        <PopoverContent className={popoverClassName ?? 'w-64'}>
          <div
            onKeyDown={(e) => {
              if (
                onApply &&
                e.key === 'Enter' &&
                e.target instanceof HTMLInputElement
              ) {
                e.preventDefault()
                handleApply()
              }
            }}
          >
            {children}
            {onApply && (
              <Button
                type="button"
                size="sm"
                className="mt-4 w-full"
                onClick={handleApply}
              >
                Apply
              </Button>
            )}
          </div>
        </PopoverContent>
      </Popover>
      {showClear && (
        <button
          type="button"
          onClick={handleClear}
          className="border-l px-2 hover:bg-muted transition-colors"
          aria-label="Clear filter"
        >
          <X className="h-4 w-4" />
        </button>
      )}
    </div>
  )
}
