import { useState, useCallback } from 'react'

/*
 * Filter apply convention
 * =======================
 * Filter popovers buffer edits in local draft state and only commit to the
 * parent's onChange when the user presses Apply (or Enter in an input). Closing
 * the popover any other way (click away, Escape) discards the draft, so the
 * chip summary and the table always show what is actually applied. This
 * prevents the table from re-fetching on every keystroke/toggle while the user
 * is still specifying their filter. Simple single-select filters (e.g.
 * Assigned To, Prospect) are exempt since selecting a value is already a single
 * deliberate action.
 *
 * Use the useDraftFilter hook below to implement this pattern. Spread its
 * chipProps onto FilterChip and use draft/setDraft instead of value/onChange
 * inside the popover body.
 */

export interface DraftFilterChipProps {
  /** Resets the draft from the committed value when the popover opens. */
  onOpenChange: (open: boolean) => void
  /** Commits the draft. FilterChip shows an Apply button when this is set. */
  onApply: () => void
}

/**
 * Hook that buffers filter edits locally and commits them on apply.
 * Returns [draft, setDraft, chipProps] — spread chipProps onto FilterChip.
 */
export function useDraftFilter<T>(
  value: T,
  onChange: (value: T) => void,
): [T, (v: T) => void, DraftFilterChipProps] {
  const [draft, setDraft] = useState(value)

  const onOpenChange = useCallback(
    (open: boolean) => {
      if (open) setDraft(value)
    },
    [value],
  )

  const onApply = useCallback(() => {
    if (draft !== value) onChange(draft)
  }, [draft, value, onChange])

  return [draft, setDraft, { onOpenChange, onApply }]
}
