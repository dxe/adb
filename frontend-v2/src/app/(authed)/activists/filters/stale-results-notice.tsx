'use client'

import { Button } from '@/components/ui/button'

interface StaleResultsNoticeProps {
  onRefresh: () => void
}

/**
 * Quiet line under the filter chips, shown once an edit made from this page
 * may have left rows on screen that the current filters exclude. The rows
 * aren't counted — the wording stays hedged because of it.
 */
export function StaleResultsNotice({ onRefresh }: StaleResultsNoticeProps) {
  return (
    <p className="text-sm text-muted-foreground" role="status">
      Some rows may no longer match your filters.{' '}
      <Button
        type="button"
        variant="link"
        size="sm"
        className="h-auto p-0 text-sm align-baseline"
        onClick={onRefresh}
      >
        Refresh
      </Button>
    </p>
  )
}
