# Bounded-height flex chain

A layout pattern that propagates the viewport's height down to an arbitrary descendant, so it can size to the remaining space — without measuring anything in JavaScript.

## Why

The naive solution to "make this region fill the remaining viewport space" is to measure in JS and subtract the chrome (e.g. navigation bar, footers, layout padding/margins).

That has well-known problems including layout shift issues and tight coupling to unrelated layout.

## Pattern

Each `flex` container inherits its parent's bounded height and re-exposes it to its own children. The cap propagates structurally — no javascript necessary.

**Anchor** — one place, at the top of the tree:

```html
<body class="h-dvh flex flex-col"></body>
```

`h-dvh` is the definite height the rest of the chain caps against.

**Link** — every element between the anchor and the consumer:

```html
<div class="flex-1 min-h-0 flex flex-col"></div>
```

- `flex-1` — fill the parent's bounded height.
- `min-h-0` — flex items default to `min-height: min-content`, which would refuse to shrink below content size. Setting it to `0` lets the cap actually take effect.
- `flex flex-col` — sets up the same contract for children, so the next link can opt in.

**Consumer** — any component that needs a definite parent height: an internally-scrolling region, a virtualized list, a canvas or widget filling its parent.

## Rules

1. **Every link must be set up correctly.** If any ancestor between the anchor and the consumer is missing `flex flex-col` or its own `flex-1 min-h-0`, the chain breaks and the consumer's `h-full` / `flex-1` resolves against `auto` — i.e. collapses to content.
2. **Opt-in is per-subtree.** Subtrees that don't need the bounded height can ignore the pattern; an `overflow-y-auto` higher in the chain is the fallback page scroller for those subtrees.
3. **Responsive variants are fine.** `md:flex-1 md:min-h-0` opts in only at `md`+, useful when mobile renders a card layout that should flow naturally.

## Pitfalls

- **Missing `min-h-0`.** Without this, a `flex-1` element won't shrink below its content.
- **A `display: block` ancestor.** Breaks the chain — `flex-1` is meaningless without a flex parent.

## Where it's used

```sh
grep -rn "bounded-height-flex-chain.md" frontend-v2/src
```
