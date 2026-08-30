'use client'

import {
  EyeOff,
  GitMerge,
  Mail,
  MoreHorizontal,
  Phone,
  Plus,
} from 'lucide-react'
import { ActivistJSON } from '@/lib/api'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { isDateAfterNMonthsAgo } from '../date-time'
import { getActivistDisplayName } from '../display-name'
import { LinkedValue } from '../linked-value'

const ACTIVE_WINDOW_MONTHS = 3

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
  const displayName = getActivistDisplayName(activist)
  const isActive =
    !!activist.last_event &&
    isDateAfterNMonthsAgo(activist.last_event, ACTIVE_WINDOW_MONTHS)

  return (
    <header className="flex flex-col gap-3">
      {/* Wraps rather than clipping: on a narrow phone the name plus both
          actions are wider than the screen, so the actions drop to their own
          row instead of pushing the overflow menu off the right edge. */}
      <div className="flex flex-wrap items-start justify-between gap-x-4 gap-y-3">
        <div className="min-w-0">
          <h1
            className={`text-3xl font-bold ${
              displayName.isPlaceholder ? 'italic text-muted-foreground' : ''
            }`}
          >
            {displayName.text}
          </h1>
          {activist.activist_level && (
            <p className="text-sm text-muted-foreground">
              {activist.activist_level}
            </p>
          )}
        </div>
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
      </div>

      {(activist.phone || activist.email) && (
        <div className="flex flex-wrap items-center gap-x-8 gap-y-1 text-sm">
          {activist.phone && (
            <span className="flex items-center gap-2">
              <Phone className="h-4 w-4 text-muted-foreground" />
              <LinkedValue value={activist.phone} linkType="tel" />
            </span>
          )}
          {activist.email && (
            <span className="flex items-center gap-2">
              <Mail className="h-4 w-4 text-muted-foreground" />
              <LinkedValue value={activist.email} linkType="mailto" />
            </span>
          )}
        </div>
      )}

      {(isActive || activist.hiatus) && (
        <div className="flex flex-wrap items-center gap-2">
          {isActive && (
            <Badge className="border-transparent bg-green-100 text-green-700 hover:bg-green-100">
              Active
            </Badge>
          )}
          {activist.hiatus && (
            <Badge className="border-transparent bg-amber-100 text-amber-700 hover:bg-amber-100">
              Hiatus
            </Badge>
          )}
        </div>
      )}
    </header>
  )
}
