'use client'

import { useCallback, useEffect, useMemo, useRef } from 'react'
import type { PointerEvent as ReactPointerEvent } from 'react'

const DEFAULT_DELAY_MS = 450
// How far the pointer may drift before the press is treated as a scroll/drag.
const MOVE_TOLERANCE_PX = 10

export interface LongPressHandlers {
  onPointerDown: (e: ReactPointerEvent) => void
  onPointerMove: (e: ReactPointerEvent) => void
  onPointerUp: () => void
  onPointerCancel: () => void
  onPointerLeave: () => void
  onContextMenu: (e: { preventDefault: () => void }) => void
}

export interface UseLongPressResult {
  handlers: LongPressHandlers
  /**
   * Whether the press that just ended was a long press, clearing the flag so a
   * later click doesn't see it again. Call this from `onClick` to suppress the
   * click that follows a long press (e.g. link navigation).
   */
  consumeLongPress: () => boolean
}

/**
 * Fires `onLongPress` when the pointer is held still on an element for
 * `delayMs`. Scrolling, dragging past MOVE_TOLERANCE_PX, or releasing early
 * cancels it.
 */
export function useLongPress(
  onLongPress: () => void,
  delayMs: number = DEFAULT_DELAY_MS,
): UseLongPressResult {
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const originRef = useRef<{ x: number; y: number } | null>(null)
  const firedRef = useRef(false)
  const onLongPressRef = useRef(onLongPress)

  useEffect(() => {
    onLongPressRef.current = onLongPress
  }, [onLongPress])

  const cancel = useCallback(() => {
    if (timerRef.current !== null) {
      clearTimeout(timerRef.current)
      timerRef.current = null
    }
    originRef.current = null
  }, [])

  // Clear a pending timer if the element unmounts mid-press.
  useEffect(() => cancel, [cancel])

  const handlers = useMemo<LongPressHandlers>(
    () => ({
      onPointerDown: (e: ReactPointerEvent) => {
        // Ignore right/middle clicks; they never become a selection gesture.
        if (e.button !== 0) return
        cancel()
        firedRef.current = false
        originRef.current = { x: e.clientX, y: e.clientY }
        timerRef.current = setTimeout(() => {
          timerRef.current = null
          firedRef.current = true
          // Short buzz on devices that support it, matching the OS convention
          // for entering a selection mode.
          navigator.vibrate?.(15)
          onLongPressRef.current()
        }, delayMs)
      },
      onPointerMove: (e: ReactPointerEvent) => {
        const origin = originRef.current
        if (!origin) return
        if (
          Math.abs(e.clientX - origin.x) > MOVE_TOLERANCE_PX ||
          Math.abs(e.clientY - origin.y) > MOVE_TOLERANCE_PX
        ) {
          cancel()
        }
      },
      onPointerUp: cancel,
      onPointerCancel: cancel,
      onPointerLeave: cancel,
      onContextMenu: (e: { preventDefault: () => void }) => {
        // Long-pressing a link opens the platform context menu on top of the
        // selection we just made, so suppress it while the gesture is live.
        if (firedRef.current || timerRef.current !== null) e.preventDefault()
      },
    }),
    [cancel, delayMs],
  )

  const consumeLongPress = useCallback(() => {
    const fired = firedRef.current
    firedRef.current = false
    return fired
  }, [])

  return { handlers, consumeLongPress }
}
