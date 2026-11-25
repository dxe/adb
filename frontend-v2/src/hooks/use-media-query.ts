'use client'

import { useCallback, useSyncExternalStore } from 'react'

/**
 * Tracks whether a CSS media query currently matches, updating on viewport
 * resize. Returns false during SSR / hydration.
 */
export function useMediaQuery(query: string): boolean {
  const subscribe = useCallback(
    (onChange: () => void) => {
      const mql = window.matchMedia(query)
      mql.addEventListener('change', onChange)
      return () => mql.removeEventListener('change', onChange)
    },
    [query],
  )
  return useSyncExternalStore(
    subscribe,
    () => window.matchMedia(query).matches,
    () => false,
  )
}
