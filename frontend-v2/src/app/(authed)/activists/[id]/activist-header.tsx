'use client'

import { EyeOff, GitMerge, Mail, MoreHorizontal, Phone } from 'lucide-react'
import { ActivistJSON } from '@/lib/api'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { isDateWithinLastMonths } from '../date-time'
import { getActivistDisplayName } from '../display-name'
import { LinkedValue } from '../linked-value'

const ACTIVE_WINDOW_MONTHS = 3

export function ActivistHeader({
  activist,
  onMerge,
  onHide,
}: {
  activist: ActivistJSON
  onMerge: () => void
  onHide: () => void
}) {
  const displayName = getActivistDisplayName(activist)
  const isActive =
    !!activist.last_event &&
    isDateWithinLastMonths(activist.last_event, ACTIVE_WINDOW_MONTHS)

  return (
    <header className="flex flex-col gap-3">
      <div className="flex items-start justify-between gap-4">
        <div>
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
