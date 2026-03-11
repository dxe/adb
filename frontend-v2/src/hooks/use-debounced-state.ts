import { useEffect, useRef, useState } from 'react'
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

  useEffect(() => {
    onChangeRef.current = onChange
  }, [onChange])

  const debouncer = useDebouncer(
    (value: string) => {
      setLastSentValue(value)
      onChangeRef.current(value)
    },
    { wait },
  )

  useEffect(() => () => debouncer.cancel(), [debouncer])

  useEffect(() => {
    if (cancelVersion > 0) {
      debouncer.cancel()
    }
  }, [cancelVersion, debouncer])

  if (externalValue !== prevExternalValue) {
    setPrevExternalValue(externalValue)
    if (externalValue === lastSentValue) {
      setLastSentValue(null)
    } else {
      setLocalValue(externalValue)
      setCancelVersion((version) => version + 1)
    }
  }

  return [
    localValue,
    (value: string) => {
      setLocalValue(value)
      debouncer.maybeExecute(value)
    },
  ]
}
