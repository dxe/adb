import { CircleHelp } from 'lucide-react'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'

export function FieldDescriptionPopover({
  label,
  description,
}: {
  label: string
  description: string
}) {
  return (
    <Popover>
      <PopoverTrigger asChild>
        <button
          type="button"
          aria-label={`About ${label}`}
          // Stop propagation so clicks don't toggle a parent <Label>'s input
          // (e.g. when the popover sits next to a checkbox label).
          onClick={(e) => e.stopPropagation()}
          className="inline-flex items-center"
        >
          <CircleHelp className="h-3.5 w-3.5 text-muted-foreground" />
        </button>
      </PopoverTrigger>
      <PopoverContent className="w-64 p-2 text-xs" side="top" align="start">
        {description}
      </PopoverContent>
    </Popover>
  )
}
