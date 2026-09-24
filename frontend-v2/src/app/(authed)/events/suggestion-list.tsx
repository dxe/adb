import { RefObject, useCallback } from 'react'
import { PopoverContent } from '@/components/ui/popover'
import { cn } from '@/lib/utils'

const MAX_HEIGHT = 300
const SIDE_OFFSET = 4
const COLLISION_PADDING = 8

type Props = {
  listboxId: string
  anchorRef: RefObject<HTMLElement | null>
  suggestions: string[]
  selectedIndex: number
  onSelect: (value: string) => void
  size?: 'sm' | 'base'
}

export function SuggestionList({
  listboxId,
  anchorRef,
  suggestions,
  selectedIndex,
  onSelect,
  size = 'base',
}: Props) {
  // On open, scroll the page just enough to fit the list below the input.
  const scrollToFit = useCallback(
    (list: HTMLUListElement | null) => {
      const anchor = anchorRef.current
      if (!list || !anchor) return
      const listHeight = Math.min(list.scrollHeight, MAX_HEIGHT)
      anchor.style.scrollMarginBottom = `${listHeight + SIDE_OFFSET + COLLISION_PADDING}px`
      anchor.scrollIntoView({ block: 'nearest' })
      anchor.style.scrollMarginBottom = ''
    },
    [anchorRef],
  )

  return (
    <PopoverContent
      className="p-0 w-[var(--radix-popover-trigger-width)]"
      side="bottom"
      align="start"
      sideOffset={SIDE_OFFSET}
      // Never flip above the input; shrink to the space below instead.
      avoidCollisions={false}
      collisionPadding={COLLISION_PADDING}
      onOpenAutoFocus={(e) => e.preventDefault()}
      onCloseAutoFocus={(e) => e.preventDefault()}
    >
      <ul
        ref={scrollToFit}
        id={listboxId}
        role="listbox"
        style={{
          maxHeight: `min(${MAX_HEIGHT}px, var(--radix-popover-content-available-height))`,
        }}
        className="overflow-y-auto rounded-md border border-gray-200 bg-white shadow-lg"
      >
        {suggestions.map((suggestion, i) => (
          <li
            key={suggestion}
            id={`${listboxId}-option-${i}`}
            role="option"
            aria-selected={i === selectedIndex}
            className={cn(
              'cursor-pointer px-3 py-1 hover:bg-gray-100',
              size === 'sm' ? 'text-sm' : 'text-base',
              i === selectedIndex ? 'bg-neutral-100' : '',
            )}
            onMouseDown={(e) => {
              e.preventDefault() // Prevents input blur from firing
              onSelect(suggestion)
            }}
          >
            {suggestion}
          </li>
        ))}
      </ul>
    </PopoverContent>
  )
}
