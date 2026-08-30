'use client'

import { EyeOff, GitMerge, MoreHorizontal, Plus } from 'lucide-react'
import { ActivistJSON } from '@/lib/api'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { ActivistIdentity } from './activist-identity'

export function ActivistHeader({
  activist,
  onLogInteraction,
  onMerge,
  onHide,
}: {
  activist: ActivistJSON
  onLogInteraction: () => void
  onMerge: () => void
  onHide: () => void
}) {
  return (
    <header>
      <ActivistIdentity
        activist={activist}
        actions={
          <div className="flex shrink-0 items-center gap-2">
            <Button type="button" onClick={onLogInteraction}>
              <Plus className="h-4 w-4" />
              Log interaction
            </Button>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button
                  type="button"
                  variant="outline"
                  size="icon"
                  aria-label="Activist actions"
                >
                  <MoreHorizontal className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem onSelect={onMerge}>
                  <GitMerge className="h-4 w-4" />
                  Merge
                </DropdownMenuItem>
                <DropdownMenuItem onSelect={onHide}>
                  <EyeOff className="h-4 w-4" />
                  Hide
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        }
      />
    </header>
  )
}
