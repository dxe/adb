'use client'

import Link from 'next/link'
import {
  ComponentPropsWithoutRef,
  FocusEvent,
  MouseEvent,
  useState,
} from 'react'

type IntentPrefetchLinkProps = Omit<
  ComponentPropsWithoutRef<typeof Link>,
  'prefetch'
>

// Use this in long lists of links to avoid viewport-triggered prefetch floods.
// Prefetch is enabled only after hover/focus.
export function IntentPrefetchLink({
  onMouseEnter,
  onFocus,
  ...props
}: IntentPrefetchLinkProps) {
  const [prefetch, setPrefetch] = useState(false)

  const handleMouseEnter = (event: MouseEvent<HTMLAnchorElement>) => {
    setPrefetch(true)
    onMouseEnter?.(event)
  }

  const handleFocus = (event: FocusEvent<HTMLAnchorElement>) => {
    setPrefetch(true)
    onFocus?.(event)
  }

  return (
    <Link
      {...props}
      prefetch={prefetch}
      onMouseEnter={handleMouseEnter}
      onFocus={handleFocus}
    />
  )
}
