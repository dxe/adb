# Components/UI

Components in this directory are generated via [shadcn](https://ui.shadcn.com/docs).

## Customizations

When upgrading shadcn components, preserve the following customizations:

### date-picker.tsx

Not a shadcn component — built from scratch using [`react-aria-components`](https://react-aria.adobe.com/react-aria/DatePicker.html) and [`@internationalized/date`](https://react-spectrum.adobe.com/internationalized/date/index.html).

React Aria's `DateInput` / `DateSegment` give us native segment editing (month, day, year each independently focusable; arrow keys increment; typing digits auto-advances) without any custom input-mask logic.

The component accepts/returns plain JS `Date` values. Two small helpers convert to/from React Aria's `CalendarDate` type internally:

```ts
toCalendarDate(date: Date): CalendarDate  // new CalendarDate(y, m+1, d)
toJSDate(date: CalendarDate): Date        // new Date(y, m-1, d)
```

**Styling:** all components are unstyled by default and take a `className` prop (or render-prop function). Styles mirror our existing `Input` tokens:
- `Group` container: `border-input`, `hover:border-gray-400`, `focus-within:border-primary focus-within:ring-1 focus-within:ring-ring`
- `DateSegment`: `data-[focused]` → `bg-primary text-primary-foreground`; placeholder slots → `text-muted-foreground`
- `CalendarCell`: `bg-primary` when selected, `bg-accent` for today/hover

**Accessibility:** `aria-label` props are required on `AriaDatePicker`, `Calendar`, and the icon-only buttons because React Aria enforces accessible names and will warn in the console without them.

### select.tsx

SelectTrigger includes custom border styling for improved visual feedback:

- `hover:border-gray-400` - Border color on hover
- `focus:border-primary` - Border color when focused
- `focus:hover:border-primary` - Maintains primary border when both focused and hovered
