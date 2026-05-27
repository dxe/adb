'use client'

import { useState } from 'react'
import { Bug, ChevronDown } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { apiClient, QueryActivistOptions } from '@/lib/api'

interface DebugMenuProps {
  queryOptions: QueryActivistOptions
}

/**
 * Render a "Debug" dropdown that can trigger logging of the current page SQL query.
 *
 * Displays a small "Debug" button which opens a popover containing a "Log page SQL query" action.
 * When the action is invoked, the component calls the API to create a debug query and alerts the returned debug id or an error message.
 *
 * @param queryOptions - Options used to construct the debug query sent to the API
 * @returns A React element containing the Debug popover and its actions
 */
export function DebugMenu({ queryOptions }: DebugMenuProps) {
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
            onClick={async () => {
              setIsOpen(false)
              try {
                const { id } = await apiClient.debugQueryActivists(queryOptions)
                window.alert(`Debug query id: ${id}`)
              } catch (err) {
                window.alert(
                  err instanceof Error
                    ? `Failed to log debug query: ${err.message}`
                    : 'Failed to log debug query.',
                )
              }
            }}
          >
            Log page SQL query
          </button>
        </div>
      </PopoverContent>
    </Popover>
  )
}
