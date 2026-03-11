'use client'

import { useEffect, useRef } from 'react'

type InfiniteScrollTriggerProps = {
  onLoadMore: () => Promise<unknown> | void
  isLoading: boolean
  canLoadMore: boolean
  rootMargin?: string
  loadingLabel?: string
  className?: string
}

export function InfiniteScrollTrigger({
  onLoadMore,
  isLoading,
  canLoadMore,
  rootMargin = '200px',
  loadingLabel = '',
  className = 'flex items-center justify-center py-4 text-sm text-muted-foreground',
}: InfiniteScrollTriggerProps) {
  const ref = useRef<HTMLDivElement>(null)
  const inFlightRef = useRef(false)

  useEffect(() => {
    const el = ref.current
    if (!el || isLoading || !canLoadMore) return

    const observer = new IntersectionObserver(
      ([entry]) => {
        // Prevent duplicate calls before React re-renders with updated loading state.
        if (!entry.isIntersecting || inFlightRef.current) return
        inFlightRef.current = true
        void Promise.resolve(onLoadMore()).finally(() => {
          inFlightRef.current = false
        })
      },
      { rootMargin },
    )

    observer.observe(el)
    return () => observer.disconnect()
  }, [canLoadMore, isLoading, onLoadMore, rootMargin])

  return (
    <div ref={ref} className={className}>
      <span role="status" aria-live="polite">
        {isLoading ? loadingLabel : ''}
      </span>
    </div>
  )
}
