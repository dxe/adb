import { useState, useMemo, useCallback } from 'react'
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

  // Sync from external changes (e.g. reset)
  if (externalValue !== prevExternal) {
    setPrevExternal(externalValue)
    setLocalValue(externalValue)
  }

  const debouncedOnChange = useMemo(
    () => liteDebounce((v: string) => onChange(v), { wait }),
    [onChange, wait],
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
