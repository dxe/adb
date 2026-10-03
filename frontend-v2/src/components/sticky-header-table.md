# StickyHeaderTable

Notes on why [`sticky-header-table.tsx`](./sticky-header-table.tsx) is built the
way it is. The height contract it depends on is documented separately in
[docs/patterns/bounded-height-flex-chain.md](../../docs/patterns/bounded-height-flex-chain.md).

## Native scrollbar, not Radix `<ScrollArea>`

This was originally a Radix `<ScrollArea>`. That component hides the native
scrollbar and draws a substitute `<div>` in its place, positioning it from a
main-thread `requestAnimationFrame` loop.

The content itself scrolls on the compositor thread, independently of the main
thread. So when the main thread is busy, the two come apart: the content keeps
scrolling smoothly while the fake thumb sits frozen at whatever offset JS last
managed to write, unsticking only once a frame is free again. The activists
table renders every loaded row unvirtualized, which is more than enough work to
make that visible.

A native scrollbar is drawn by the browser in lockstep with the scroll, so it
cannot drift no matter how busy we are.

This is a property of the page, not of `<ScrollArea>` — the substitute thumb
keeps up fine when the main thread is idle. A short dialog body or a dropdown
list has no reason to avoid it. Reach for a native scroller when the scrolling
region shares a page with expensive work.

The styling `<ScrollArea>` provided is replaced by `scrollbar-width` and
`scrollbar-color`, which are native CSS.

## `min-h-0` on the scroll container

Flex items default to `min-height: min-content`, which refuses to shrink below
content size. Setting it to `0` is what lets flex-shrink cap the container at
the parent's height when the table is taller — i.e. what makes it scroll
internally rather than grow.

## Handing scrolling back to the page under `max-height: 700px`

At small browser heights an internally-scrolling region is cramped, and the
better UX is to let the whole page scroll instead. The `[@media(max-height:700px)]`
overrides switch the container out of the way rather than resizing it:

- `overflow-visible` stops it being a scrolling ancestor at all, so the sticky
  header resolves against whatever scroll container is next up the tree.
- `min-h-[auto]` reverts the flex-shrink cap above, so it grows to content size.
- `w-max max-w-none` does the same horizontally. Without it the wrapper stays at
  the consumer's width while the inner table keeps its explicit wider width,
  leaving rows visibly spilling past the right border. Overflow then propagates
  up into horizontal page scroll.

## `[*:has(>&)]:overflow-visible` on the `<Table>`

shadcn's `<Table>` wraps its `<table>` in a `<div className="overflow-auto">`.
That wrapper would otherwise be the nearest scrolling ancestor, and the sticky
header would pin against it instead of our container.

`[*:has(>&)]` selects the parent of the element the class is on — i.e. shadcn's
wrapper — and flips it back to `overflow: visible`. Targeting it from the child
this way means `ui/table.tsx` stays unpatched and regenerable with shadcn.
