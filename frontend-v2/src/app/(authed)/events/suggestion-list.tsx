import { PopoverContent } from '@/components/ui/popover'
import { cn } from '@/lib/utils'

type Props = {
  suggestions: string[]
  selectedIndex: number
  onSelect: (value: string) => void
  size?: 'sm' | 'base'
}

export function SuggestionList({
  suggestions,
  selectedIndex,
  onSelect,
  size = 'base',
}: Props) {
  return (
    <PopoverContent
      className="p-0 w-[var(--radix-popover-trigger-width)]"
      align="start"
      sideOffset={4}
      onOpenAutoFocus={(e) => e.preventDefault()}
      onCloseAutoFocus={(e) => e.preventDefault()}
    >
      <ul className="max-h-[300px] overflow-y-auto rounded-md border border-gray-200 bg-white shadow-lg">
        {suggestions.map((suggestion, i) => (
          <li
            key={suggestion}
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
