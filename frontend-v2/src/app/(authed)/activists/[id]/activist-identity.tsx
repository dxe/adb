'use client'

import { ReactNode } from 'react'
import { Mail, Phone } from 'lucide-react'
import { ActivistJSON } from '@/lib/api'
import { Badge } from '@/components/ui/badge'
import { isDateAfterNMonthsAgo } from '../date-time'
import { getActivistDisplayName } from '../display-name'
import { LinkedValue } from '../linked-value'

const ACTIVE_WINDOW_MONTHS = 3

type Variant = 'page' | 'compact'

// Who this activist is: name, level, contact links and status chips.
export function ActivistIdentity({
  activist,
  variant = 'page',
  actions,
}: {
  activist: ActivistJSON
  // 'page' is the document heading; 'compact' sits under a heading that is
  // already on screen (the dialog's title), so it is quieter and tighter.
  variant?: Variant
  actions?: ReactNode
}) {
  const displayName = getActivistDisplayName(activist)
  const isActive =
    !!activist.last_event &&
    isDateAfterNMonthsAgo(activist.last_event, ACTIVE_WINDOW_MONTHS)
  const isPage = variant === 'page'
  const Heading = isPage ? 'h1' : 'h2'

  return (
    <div className={`flex flex-col ${isPage ? 'gap-3' : 'gap-2'}`}>
      <div className="flex flex-wrap items-start justify-between gap-x-4 gap-y-3">
        <div className="min-w-0">
          <Heading
            className={`${isPage ? 'text-3xl font-bold' : 'text-lg font-semibold'} ${
              displayName.isPlaceholder ? 'italic text-muted-foreground' : ''
            }`}
          >
            {displayName.text}
          </Heading>
          {activist.activist_level && (
            <p className="text-sm text-muted-foreground">
              {activist.activist_level}
            </p>
          )}
        </div>
        {actions}
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
            <span className="flex min-w-0 items-center gap-2">
              <Mail className="h-4 w-4 shrink-0 text-muted-foreground" />
              <span className="truncate">
                <LinkedValue value={activist.email} linkType="mailto" />
              </span>
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
    </div>
  )
}
