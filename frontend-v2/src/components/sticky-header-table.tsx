'use client'

import * as React from 'react'
import { cn } from '@/lib/utils'
import { Table } from '@/components/ui/table'

interface StickyHeaderTableProps extends React.ComponentProps<typeof Table> {
  /** Rendered inside the scroll viewport below the table, e.g. an infinite-scroll trigger. */
  footer?: React.ReactNode
}

/**
 * <Table> wrapped in a scroll container so a sticky header — set by the
 * consumer with e.g. `<TableHeader className="sticky top-0 z-10 bg-background">`
 * — pins against the scroll viewport instead of scrolling with the page.
 *
 * Short tables size to their content; tall tables cap at the parent's
 * height and scroll internally, so they don't push the page.
 *
 * Usage requirement: must be rendered as the consumer of a bounded-height
 * flex chain — i.e. inside a `<div className="flex flex-col flex-1 min-h-0">`
 * whose ancestors all opt in. Without that, the viewport has no definite
 * height to cap against and the sticky header has no scroll range to pin to.
 * See frontend-v2/docs/patterns/bounded-height-flex-chain.md
 *
 * Why the scroll container is built the way it is — the native scrollbar, the
 * `min-h-0`, the `max-height:700px` overrides — is in ./sticky-header-table.md
 */
export function StickyHeaderTable({
  footer,
  className,
  children,
  ...tableProps
}: StickyHeaderTableProps) {
  return (
    // `min-h-0` lets flex-shrink cap this at the wrapper's height; the
    // max-height overrides hand scrolling back to the page on short viewports.
    // For details, see ./sticky-header-table.md
    <div
      className={cn(
        'min-h-0 max-w-full overflow-auto rounded-md border',
        '[scrollbar-width:thin] [scrollbar-color:hsl(var(--border))_transparent]',
        '[@media(max-height:700px)]:min-h-[auto] [@media(max-height:700px)]:overflow-visible',
        '[@media(max-height:700px)]:w-max [@media(max-height:700px)]:max-w-none',
      )}
    >
      {/* `[*:has(>&)]` targets shadcn's <Table> wrapper div, flipping its
          `overflow-auto` back to visible so it isn't the sticky scroll container.
          For details, see ./sticky-header-table.md */}
      <Table
        className={cn('[*:has(>&)]:overflow-visible', className)}
        {...tableProps}
      >
        {children}
      </Table>
      {footer}
    </div>
  )
}
