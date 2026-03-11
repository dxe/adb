import { useEffect, useLayoutEffect, useRef, useState } from 'react'
import { useDebouncer } from '@tanstack/react-pacer'

/**
 * Local state that stays in sync with an external value and debounces writes.
 * External resets cancel any pending debounce unless they are the echo of our
 * own debounced update.
 */
export function useDebouncedState(
  externalValue: string,
  onChange: (v: string) => void,
  { wait = 300 }: { wait?: number } = {},
): [string, (v: string) => void] {
  const [localValue, setLocalValue] = useState(externalValue)
  const [prevExternalValue, setPrevExternalValue] = useState(externalValue)
  const [lastSentValue, setLastSentValue] = useState<string | null>(null)
  const [cancelVersion, setCancelVersion] = useState(0)
  const onChangeRef = useRef(onChange)
  // Tracks the latest scheduled debounced value so stale callbacks can no-op.
  const pendingValueRef = useRef<string | null>(null)

  useLayoutEffect(() => {
    onChangeRef.current = onChange
  }, [onChange])

  const debouncer = useDebouncer(
    (value: string) => {
      if (pendingValueRef.current !== value) {
        return
      }

      pendingValueRef.current = null
      setLastSentValue(value)
      onChangeRef.current(value)
    },
    { wait },
  )

  useEffect(() => () => debouncer.cancel(), [debouncer])

  useLayoutEffect(() => {
    if (cancelVersion > 0) {
      pendingValueRef.current = null
      debouncer.cancel()
    }
  }, [cancelVersion, debouncer])

  if (externalValue !== prevExternalValue) {
    setPrevExternalValue(externalValue)
    setLastSentValue(null)
    if (externalValue !== lastSentValue) {
      setLocalValue(externalValue)
      setCancelVersion((version) => version + 1)
    }
  }

  return [
    localValue,
    (value: string) => {
      setLocalValue(value)
      pendingValueRef.current = value
      debouncer.maybeExecute(value)
    },
  ]
}
