import { useState, useMemo, useCallback, useRef } from 'react'
import { liteDebounce } from '@tanstack/pacer-lite'

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

  // Track the last value we actually sent so we can distinguish our own
  // debounce-triggered URL updates from genuinely external changes (e.g. reset).
  const lastSentRef = useRef(externalValue)

  // Sync from external changes (e.g. reset), but not when the URL changed
  // because of our own debounce firing (which would override in-progress typing).
  if (externalValue !== prevExternal) {
    setPrevExternal(externalValue)
    if (externalValue !== lastSentRef.current) {
      setLocalValue(externalValue)
    }
  }

  const debouncedOnChange = useMemo(
    () =>
      liteDebounce(
        (v: string) => {
          lastSentRef.current = v
          onChangeRef.current(v)
        },
        { wait },
      ),
    [wait],
  )

  const handleChange = useCallback(
    (v: string) => {
      setLocalValue(v)
      debouncedOnChange(v)
    },
    [debouncedOnChange],
  )

  return [localValue, handleChange]
}
