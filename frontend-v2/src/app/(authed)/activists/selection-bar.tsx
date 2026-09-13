'use client'

import { X } from 'lucide-react'
import { Button } from '@/components/ui/button'

interface SelectionBarProps {
  count: number
  onAssign: () => void
  onClear: () => void
}

/**
 * Floating bar shown while activists are selected. Fixed to the bottom of the
 * viewport so it stays reachable while scrolling a long list.
 */
export function SelectionBar({ count, onAssign, onClear }: SelectionBarProps) {
  if (count === 0) return null

  return (
    // The outer layer spans the viewport but ignores pointer events so it
    // never blocks the rows underneath it.
    <div
      className="pointer-events-none fixed inset-x-0 bottom-0 z-40 flex justify-center p-4"
      style={{ paddingBottom: 'calc(1rem + env(safe-area-inset-bottom))' }}
    >
      <div className="pointer-events-auto flex items-center gap-3 rounded-full border bg-background/95 py-2 pl-5 pr-2 shadow-lg backdrop-blur">
        <span
          className="text-sm font-medium whitespace-nowrap"
          aria-live="polite"
        >
          {count} selected
        </span>
        <Button type="button" size="sm" onClick={onAssign}>
          Assign to…
        </Button>
        <Button
          type="button"
          size="icon"
          variant="ghost"
          className="h-8 w-8 rounded-full"
          onClick={onClear}
          aria-label="Clear selection"
        >
          <X className="h-4 w-4" />
        </Button>
      </div>
    </div>
  )
}
