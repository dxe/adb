import type { Metadata, Viewport } from 'next'
import { Toaster } from 'react-hot-toast'
import Providers from './providers'
import { cn } from '@/lib/utils'
import '@/styles/globals.css'
// TODO(jh): hopefully these buefy styles don't conflict w/
// tailwind... can we maybe scope it to the nav.tsx somehow?
import '../../../frontend/static/external/buefy.min.css'

export const metadata: Metadata = {
  title: 'Activist Database',
  description: 'Activist Database for Direct Action Everywhere',
  icons: '/v2/favicon.png',
}

export const viewport: Viewport = {
  width: 'device-width',
  initialScale: 1,
  maximumScale: 1,
  userScalable: false,
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <head>
        <meta charSet="utf-8" />
      </head>
      <body
        className={cn(
          'antialiased',
          // Prevents layout shift since we are still using buefy styles for the nav.
          // This should be removed once the nav is no longer using buefy.
          'has-navbar-fixed-top',
        )}
      >
        <Providers>{children}</Providers>
        <Toaster position="bottom-right" />
      </body>
    </html>
  )
}
