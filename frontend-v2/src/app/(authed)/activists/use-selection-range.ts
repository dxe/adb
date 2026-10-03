'use client'

import { useCallback, useMemo, useRef, type MouseEvent } from 'react'
import type { CheckedState } from '@radix-ui/react-checkbox'
import type { ActivistSelection } from './use-activist-selection'

/** Props to spread onto a row's selection checkbox. */
export interface RowCheckboxProps {
  onClick: (event: MouseEvent) => void
  onMouseDown: (event: MouseEvent) => void
  onCheckedChange: (checked: CheckedState) => void
}

export interface SelectionRange {
  rowCheckboxProps: (id: number) => RowCheckboxProps
  /** For the header checkbox, which selects or deselects every loaded row. */
  onAllCheckedChange: (checked: CheckedState) => void
}

/**
 * Shift-click range selection over the rows on screen.
 *
 * A plain click both toggles its row and becomes the anchor. A shift-click
 * applies whatever it did to its own checkbox — check or uncheck — to every
 * row between the anchor and itself, then becomes the anchor in turn.
 *
 * `rowIds` gives the on-screen order the range is taken along; an anchor that
 * is no longer among them (the list was refiltered, say) is ignored.
 */
export function useSelectionRange(
  rowIds: number[],
  selection: ActivistSelection | undefined,
): SelectionRange {
  // The row the user last checked or unchecked, and the end the next
  // shift-click extends from.
  const anchorIdRef = useRef<number | null>(null)
  // Radix hands `onCheckedChange` no event, so the modifier is stashed from
  // the click that is about to produce it.
  const extendRef = useRef(false)

  const rowCheckboxProps = useCallback(
    (id: number): RowCheckboxProps => ({
      onClick: (event) => {
        extendRef.current = event.shiftKey
      },
      // Shift-clicking otherwise extends the document text selection across
      // the rows in between, highlighting the whole range.
      onMouseDown: (event) => {
        if (event.shiftKey) event.preventDefault()
      },
      onCheckedChange: (checked) => {
        const extend = extendRef.current
        extendRef.current = false
        if (!selection) return

        const anchorId = anchorIdRef.current
        anchorIdRef.current = id

        const anchorIndex = anchorId === null ? -1 : rowIds.indexOf(anchorId)
        const index = rowIds.indexOf(id)
        if (!extend || anchorIndex === -1 || index === -1) {
          selection.onToggle(id)
          return
        }

        const from = Math.min(anchorIndex, index)
        const to = Math.max(anchorIndex, index)
        selection.onSetMany(rowIds.slice(from, to + 1), checked === true)
      },
    }),
    [rowIds, selection],
  )

  const onAllCheckedChange = useCallback(
    (checked: CheckedState) => {
      // Select-all leaves no meaningful end to extend from.
      anchorIdRef.current = null
      selection?.onSetMany(rowIds, checked === true)
    },
    [rowIds, selection],
  )

  return useMemo(
    () => ({ rowCheckboxProps, onAllCheckedChange }),
    [rowCheckboxProps, onAllCheckedChange],
  )
}
