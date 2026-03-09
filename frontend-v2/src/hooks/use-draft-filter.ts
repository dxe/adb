import { useState, useCallback } from 'react'

/*
 * Filter commit-on-close convention
 * ==================================
 * Filter popovers should buffer edits in local draft state and only commit to
 * the parent's onChange when the popover closes. This prevents the table from
 * re-fetching on every keystroke/toggle while the user is still specifying
 * their filter. Simple single-select filters (e.g. Assigned To, Prospect) are
 * exempt since selecting a value is already a single deliberate action.
 *
 * Use the useDraftFilter hook below to implement this pattern. Pass its
 * onOpenChange to FilterChip and use draft/setDraft instead of value/onChange
 * inside the popover body.
 */

/**
 * Hook that buffers filter edits locally and commits on popover close.
 * Returns [draft, setDraft, onOpenChange] — wire onOpenChange to FilterChip.
 */
export function useDraftFilter<T>(
  value: T,
  onChange: (value: T) => void,
): [T, (v: T) => void, (open: boolean) => void] {
  const [draft, setDraft] = useState(value)

  const onOpenChange = useCallback(
    (open: boolean) => {
      if (open) {
        setDraft(value)
      } else {
        if (draft !== value) onChange(draft)
      }
    },
    [draft, value, onChange],
  )

  return [draft, setDraft, onOpenChange]
}
