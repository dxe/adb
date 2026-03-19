'use client'

import { useEffect } from 'react'
import { usePathname } from 'next/navigation'

const MOBILE_BREAKPOINT = '(max-width: 767px)'

/**
 * Adds the site background on non-full-screen desktop pages.
 *
 * Full screen pages take the full width of the screen, but not necessarily the
 * full height if there is not enough content, which may expose the background
 * image. This prevents loading the image on those pages for aesthetic reasons.
 *
 * The image is removed for mobile users as well as they are more likely to have
 * a poor internet connection.
 */
export default function SiteBackgroundController() {
  const pathname = usePathname()

  useEffect(() => {
    const mediaQuery = window.matchMedia(MOBILE_BREAKPOINT)
    const isFullScreenPage =
      pathname === '/activists' || pathname === '/intl/organizers'

    const applyBackgroundClass = () => {
      const shouldShowBackground = !isFullScreenPage && !mediaQuery.matches
      document.body.classList.toggle('with-site-bg', shouldShowBackground)
    }

    applyBackgroundClass()
    mediaQuery.addEventListener('change', applyBackgroundClass)

    return () => {
      mediaQuery.removeEventListener('change', applyBackgroundClass)
    }
  }, [pathname])

  return null
}
