import { useState, useRef } from 'react'
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
  /** Called when the popover opens or closes. Used by useDraftFilter to
   *  sync/commit draft state. Not called when the × button is pressed. */
  onOpenChange?: (open: boolean) => void
}

export function FilterChip({
  label,
  summary,
  onClear,
  removable,
  children,
  defaultOpen,
  popoverClassName,
  onOpenChange,
}: FilterChipProps) {
  const showClear = !!summary || removable
  const [open, setOpen] = useState(defaultOpen ?? false)
  // When × is clicked we close the popover and call onClear directly.
  // The ref prevents onOpenChange from also firing (which would commit
  // stale draft state over the cleared value).
  const clearingRef = useRef(false)

  const handleOpenChange = (newOpen: boolean) => {
    setOpen(newOpen)
    if (newOpen) {
      // Always sync draft from value when opening.
      onOpenChange?.(true)
    } else if (!clearingRef.current) {
      // Commit draft on normal close, but not when × was clicked.
      onOpenChange?.(false)
    }
    clearingRef.current = false
  }

  const handleClear = () => {
    clearingRef.current = true
    setOpen(false)
    onClear()
    // If the popover was already closed, Radix won't call handleOpenChange
    // to consume clearingRef. Reset it asynchronously so it doesn't block
    // the next open.
    requestAnimationFrame(() => {
      clearingRef.current = false
    })
  }

  return (
    <div className="flex shrink-0 items-stretch rounded-md border bg-card overflow-hidden h-12">
      <Popover open={open} onOpenChange={handleOpenChange}>
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
