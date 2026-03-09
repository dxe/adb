import { useState, useMemo, useCallback, useRef, useEffect } from 'react'
import { LiteDebouncer } from '@tanstack/pacer-lite'

/**
 * Local state that stays in sync with an external value (e.g. a URL param)
 * and debounces calls to onChange. Use for text inputs that should update
 * immediately in the UI but only trigger side effects (queries, URL writes)
 * after the user stops typing.
 *
 * External changes (e.g. a filter reset) are synced back to local state via
 * the derived-state-during-render pattern, so no useEffect is needed.
 */
export function useDebouncedState(
  externalValue: string,
  onChange: (v: string) => void,
  { wait = 300 }: { wait?: number } = {},
): [string, (v: string) => void] {
  const [localValue, setLocalValue] = useState(externalValue)
  const [prevExternal, setPrevExternal] = useState(externalValue)

  // Use a ref so the debounced function is stable across renders (doesn't
  // recreate on every render when onChange is an inline function).
  const onChangeRef = useRef(onChange)
  onChangeRef.current = onChange

  // One-shot flag: set when we fire the debounced onChange so we skip exactly
  // the next external sync that echoes our own write back, without accidentally
  // dropping a genuine external change that happens to equal lastSentRef.
  const skipNextSyncRef = useRef(false)

  // Set during render when a genuine external reset is detected; read in a
  // useEffect so debouncer.cancel() runs after commit, not during render.
  const needsCancelRef = useRef(false)

  const debouncer = useMemo(
    () =>
      new LiteDebouncer(
        (v: string) => {
          skipNextSyncRef.current = true
          onChangeRef.current(v)
        },
        { wait },
      ),
    [wait],
  )

  // Cancel any pending debounce on unmount, and also when wait changes
  // (useMemo creates a new debouncer, so the effect re-runs and cancels the old one).
  useEffect(() => () => debouncer.cancel(), [debouncer])

  // Sync from external changes (e.g. reset), but not when the URL changed
  // because of our own debounce firing (which would override in-progress typing).
  if (externalValue !== prevExternal) {
    setPrevExternal(externalValue)
    if (skipNextSyncRef.current) {
      skipNextSyncRef.current = false
    } else {
      // Genuine external change: flag for cancellation (done in the effect
      // below) and update local state to match.
      needsCancelRef.current = true
      setLocalValue(externalValue)
    }
  }

  // Defer debouncer.cancel() to after commit so it isn't a side effect during
  // render. The ref acts as a one-shot signal set above.
  useEffect(() => {
    if (needsCancelRef.current) {
      needsCancelRef.current = false
      debouncer.cancel()
    }
  }, [externalValue, debouncer])

  const handleChange = useCallback(
    (v: string) => {
      setLocalValue(v)
      debouncer.maybeExecute(v)
    },
    [debouncer],
  )

  return [localValue, handleChange]
}
