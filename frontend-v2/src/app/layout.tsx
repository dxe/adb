import type { Metadata, Viewport } from 'next'
import { Toaster } from 'react-hot-toast'
import { ApiClient } from '@/lib/api'
import { getCookies } from '@/lib/auth'
import Providers from './providers'
import '@/styles/globals.css'

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

export default async function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <head>
        <meta charSet="utf-8" />
        {/* See README for notes on CSRF implementation. */}
        <meta name="csrf-token" content={await fetchCsrfToken()} />
      </head>
      {/* Top padding is to make room for the fixed navbar. */}
      <body className="antialiased">
        {/*
          Bottom padding uses inline style because Safari doesn't respect
          pb-[3.25rem] Tailwind class when combined with body's background styles.
          Inline styles work reliably across all browsers.
        */}
        <div className="pt-[3.25rem] min-h-screen" style={{ paddingBottom: '3rem' }}>
          <Providers>{children}</Providers>
        </div>
        <Toaster position="bottom-right" />
      </body>
    </html>
  )
}

async function fetchCsrfToken(): Promise<string | undefined> {
  const apiClient = new ApiClient(await getCookies())
  try {
    return await apiClient.fetchCsrfToken()
  } catch (err) {
    console.error(`Failed to preload CSRF token: ${err}`)
    return undefined
  }
}
