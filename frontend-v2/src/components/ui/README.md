# Components/UI

Components in this directory are generated via [shadcn](https://ui.shadcn.com/docs).

## Customizations

When upgrading shadcn components, preserve the following customizations:

### date-picker.tsx

Completely rewritten from the default shadcn button-trigger implementation. The original rendered a `<Button>` that opened a calendar popover; this version replaces it with a masked text input so users can type or pick a date.

Key changes:

- **Input mask** — the input always shows `MM/DD/YYYY` as a template. Digit slots are tracked internally as a fixed 8-character string (`'_'` = unfilled). `buildMasked` maps slots to the display string; `displayCursorToDigitCursor` maps display cursor positions back to slot indices for keyboard handling.
- **Segment editing** — double-clicking a segment (month/day/year) selects it; the next digit typed fills from the start of that segment and clears the rest, so adjacent segments never bleed in. The compact-string approach of the original would have caused this bleed.
- **Overwrite mode** — typing with a cursor (no selection) overwrites the slot under the cursor and advances, matching standard input-mask UX.
- **Calendar icon on the left** — positioned with `absolute left-3`; the input uses `style={{ paddingLeft: '2.5rem' }}` (inline style) rather than a Tailwind class because `@tailwindcss/forms` injects unlayered global padding on `input` elements that beats Tailwind utility classes regardless of specificity.
- **Cursor placement via `useLayoutEffect`** — a `pendingCursorRef` is set by `moveCursor()` and applied in `useLayoutEffect` (before paint) to avoid the one-frame cursor-reset flicker that `setTimeout` caused.
- **Popover wiring** — uses `PopoverAnchor` instead of `PopoverTrigger` so the input (not the icon button) acts as the anchor. `onOpenAutoFocus` and `onMouseDown` on the content prevent Radix from stealing focus from the input. `skipNextOpenRef` prevents the calendar from re-opening when focus returns after a date is selected or the icon is clicked.
- **Enter key** — closes the inner calendar without propagating to outer forms; `e.stopPropagation()` is called so a parent `FilterChip` popover can independently handle Enter to commit its own draft.

### select.tsx

SelectTrigger includes custom border styling for improved visual feedback:

- `hover:border-gray-400` - Border color on hover
- `focus:border-primary` - Border color when focused
- `focus:hover:border-primary` - Maintains primary border when both focused and hovered
