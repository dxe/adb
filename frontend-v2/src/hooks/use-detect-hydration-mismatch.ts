'use client'

import { useEffect, useRef } from 'react'

interface UseDetectHydrationMismatchOptions<T> {
  label: string
  serverValue?: T
  clientValue: T
}

export function useDetectHydrationMismatch<T>({
  label,
  serverValue,
  clientValue,
}: UseDetectHydrationMismatchOptions<T>) {
  const hasCompared = useRef(false)

  useEffect(() => {
    if (
      process.env.NODE_ENV !== 'development' ||
      serverValue === undefined ||
      hasCompared.current
    ) {
      return
    }
    hasCompared.current = true

    if (JSON.stringify(serverValue) !== JSON.stringify(clientValue)) {
      console.error(`[${label}] SSR/client mismatch`, {
        serverValue,
        clientValue,
      })
    }
  }, [clientValue, label, serverValue])
}
