'use client'

import * as React from 'react'
import { cn } from '@/lib/utils'
import { ScrollArea, ScrollBar } from '@/components/ui/scroll-area'
import { Table } from '@/components/ui/table'

interface StickyHeaderTableProps extends React.ComponentProps<typeof Table> {
  /** Rendered inside the scroll viewport below the table, e.g. an infinite-scroll trigger. */
  footer?: React.ReactNode
}

/**
 * <Table> wrapped in a <ScrollArea> so a sticky header — set by the consumer
 * with e.g. `<TableHeader className="sticky top-0 z-10 bg-background">` —
 * pins against the scroll viewport instead of scrolling with the page.
 *
 * Short tables size to their content; tall tables cap at the parent's
 * height and scroll internally, so they don't push the page.
 *
 * Usage requirement: must be rendered as the consumer of a bounded-height
 * flex chain — i.e. inside a `<div className="flex flex-col flex-1 min-h-0">`
 * whose ancestors all opt in. Without that, the viewport has no definite
 * height to cap against and the sticky header has no scroll range to pin to.
 * See frontend-v2/docs/patterns/bounded-height-flex-chain.md
 */
export function StickyHeaderTable({
  footer,
  className,
  children,
  ...tableProps
}: StickyHeaderTableProps) {
  return (
    <>
      {/*
        `min-h-0` lets flex-shrink cap the ScrollArea at the wrapper's height
        when the table is taller. Flex items default to `min-height:
        min-content`, which would block that shrink and leave the Viewport's
        `h-full` with no smaller-than-content parent to resolve against.
      */}
      <ScrollArea className="min-h-0 max-w-full rounded-md border">
        {/*
        shadcn's <Table> wraps its <table> in a `<div className="overflow-auto">`.
        That wrapper would otherwise be the sticky header's scroll container;
        `[*:has(>&)]:overflow-visible` flips it back to `overflow:visible` so
        sticky resolves against the <ScrollArea> above.
        
        (`[*:has(>&)]` targets the parent of the <table>, i.e. shadcn's wrapper
        so we don't have to patch <Table>, keeping it regenerable with shadcn.)
        */}
        <Table
          className={cn('[*:has(>&)]:overflow-visible', className)}
          {...tableProps}
        >
          {children}
        </Table>
        {footer}
        <ScrollBar orientation="horizontal" />
      </ScrollArea>
    </>
  )
}
