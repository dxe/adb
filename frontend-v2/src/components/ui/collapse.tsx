import * as React from 'react'
import { cn } from '@/lib/utils'

// Animates its children open/closed with a combined slide + fade. Stays mounted
// in both states (using the grid-rows 0fr->1fr trick) so the content can
// animate *out* as well as in, and so it grows to the content's natural height
// without needing a fixed max-height. While closed it's marked `inert` so the
// still-mounted fields are skipped by tab order and screen readers.
//
// `className` is applied to the outer (animated) element. Inside a flex-gap
// stack a collapsed item still leaves the parent's gap on both sides; pass a
// `-mt-*` matching the parent gap that's active only while closed (e.g.
// `cn(!open && '-mt-4')`) to absorb it — the negative margin transitions along
// with the slide.
export const Collapse = ({
  open,
  className,
  children,
}: {
  open: boolean
  className?: string
  children: React.ReactNode
}) => (
  <div
    className={cn(
      'grid transition-all duration-200 ease-out motion-reduce:transition-none',
      open ? 'grid-rows-[1fr] opacity-100' : 'grid-rows-[0fr] opacity-0',
      className,
    )}
    inert={!open}
  >
    {/* overflow-hidden does the vertical clipping for the slide, but it also
        clips the focus ring on fields inside. px-1/-mx-1 gives the ring room on
        the left/right without shifting the content (the negative margin cancels
        the padding). py-1/-my-1 does the same top/bottom, but only while open —
        vertical padding would otherwise keep the box ~8px tall when collapsed
        instead of fully closing. */}
    <div className={cn('overflow-hidden px-1 -mx-1', open && 'py-1 -my-1')}>
      {children}
    </div>
  </div>
)
