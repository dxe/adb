import { useState, useEffect } from 'react'
import { ActivistRegistry } from './activist-registry'

/**
 * Singleton cache for the activist registry.
 * Initialized once globally and shared across all components.
 */
let cachedRegistry: ActivistRegistry | null = null
let initPromise: Promise<ActivistRegistry> | null = null

/**
 * Custom hook to access the activist registry.
 * Uses a singleton pattern to initialize the registry once and share it across components.
 *
 * @returns Object containing the registry instance and loading state
 */
export function useActivistRegistry() {
  const [registry, setRegistry] = useState<ActivistRegistry | null>(
    cachedRegistry,
  )
  const [isLoading, setIsLoading] = useState(!cachedRegistry)

  useEffect(() => {
    // If already cached, no need to initialize
    if (cachedRegistry) {
      return
    }

    // Start initialization once globally
    if (!initPromise) {
      initPromise = ActivistRegistry.create().then((reg) => {
        cachedRegistry = reg
        return reg
      })
    }

    // Wait for initialization and update state
    initPromise.then((reg) => {
      setRegistry(reg)
      setIsLoading(false)
    })
  }, [])

  return { registry, isLoading }
}
