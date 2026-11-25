import { useState } from 'react'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { ChevronDown, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import {
  Dialog,
  DialogContent,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { useMediaQuery } from '@/hooks/use-media-query'

/** Below this viewport height a popover has too little room; use a modal. */
const SHORT_VIEWPORT_QUERY = '(max-height: 700px)'

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
  /** Empties the draft. Clear button is shown beside Apply; disabled when
   *  canReset is false. */
  onReset?: () => void
  canReset?: boolean
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
  onReset,
  canReset,
}: FilterChipProps) {
  const showClear = !!summary || removable
  const [open, setOpen] = useState(defaultOpen ?? false)
  const shortViewport = useMediaQuery(SHORT_VIEWPORT_QUERY)
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

  const trigger = (
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
  )

  // The body scrolls on its own so the Apply/Clear row stays visible when the
  // container is height-limited.
  const body = (
    <div
      className="flex min-h-0 flex-1 flex-col"
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
      {/* Negative margin + padding keeps focus rings from being clipped. */}
      <div className="-m-1 min-h-0 flex-1 overflow-y-auto p-1">{children}</div>
      {onApply && (
        <div className="mt-4 flex shrink-0 flex-wrap gap-2">
          {onReset && (
            <Button
              type="button"
              variant="outline"
              size="sm"
              className="min-w-24 flex-1"
              disabled={!canReset}
              onClick={onReset}
            >
              Clear
            </Button>
          )}
          <Button
            type="button"
            size="sm"
            className="min-w-24 flex-1"
            onClick={handleApply}
          >
            Apply
          </Button>
        </div>
      )}
    </div>
  )

  return (
    <div className="flex shrink-0 items-stretch rounded-md border bg-card overflow-hidden h-12">
      {shortViewport ? (
        // A popover only gets the space above or below the chip, which is
        // tiny on a short viewport, so use a centered modal instead.
        <Dialog open={open} onOpenChange={handleOpenChange}>
          <DialogTrigger asChild>{trigger}</DialogTrigger>
          <DialogContent
            aria-describedby={undefined}
            className="flex max-h-[calc(100dvh-1rem)] w-[calc(100vw-1rem)] max-w-sm flex-col gap-3 p-4"
          >
            <DialogTitle className="shrink-0 pr-6 text-base">
              {label}
            </DialogTitle>
            {body}
          </DialogContent>
        </Dialog>
      ) : (
        <Popover open={open} onOpenChange={handleOpenChange}>
          <PopoverTrigger asChild>{trigger}</PopoverTrigger>
          <PopoverContent
            collisionPadding={8}
            className={cn(
              'flex max-h-[var(--radix-popover-content-available-height)] flex-col overflow-hidden',
              popoverClassName ?? 'w-64',
            )}
          >
            {body}
          </PopoverContent>
        </Popover>
      )}
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
