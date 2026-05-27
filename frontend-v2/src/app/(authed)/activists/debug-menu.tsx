'use client'

import { useState } from 'react'
import { Bug, ChevronDown } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'

export function DebugMenu() {
  const [isOpen, setIsOpen] = useState(false)

  return (
    <Popover open={isOpen} onOpenChange={setIsOpen}>
      <PopoverTrigger asChild>
        <Button variant="ghost" size="sm" className="h-12 gap-1">
          <Bug className="h-4 w-4" />
          Debug
          <ChevronDown className="h-4 w-4" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-52 p-2" align="start">
        <div className="flex flex-col">
          <button
            type="button"
            className="flex w-full items-center rounded px-2 py-1.5 text-sm hover:bg-muted transition-colors text-left"
            onClick={() => {
              setIsOpen(false)
              console.log('not implemented')
            }}
          >
            Log page SQL query
          </button>
        </div>
      </PopoverContent>
    </Popover>
  )
}
