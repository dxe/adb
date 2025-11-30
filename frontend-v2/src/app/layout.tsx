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
      <body className="antialiased pt-[3.25rem]">
        <Providers>{children}</Providers>
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
