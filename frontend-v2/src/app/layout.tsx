import type { Metadata, Viewport } from 'next'
import { Toaster } from 'react-hot-toast'
import { ApiClient } from '@/lib/api'
import { getCookies } from '@/lib/auth'
import Providers from './providers'
import SiteBackgroundController from './site-background-controller'
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
    <html lang="en" data-env={process.env.NODE_ENV}>
      <head>
        <meta charSet="utf-8" />
        {/* See README for notes on CSRF implementation. */}
        <meta name="csrf-token" content={await fetchCsrfToken()} />
      </head>
      {/* Top padding is to make room for the fixed navbar. */}
      {/* Bounded-height flex chain anchor — see frontend-v2/docs/patterns/bounded-height-flex-chain.md */}
      <body className="antialiased h-dvh flex flex-col">
        <SiteBackgroundController />
        {/*
          Bounded-height flex chain link — see frontend-v2/docs/patterns/bounded-height-flex-chain.md
          `overflow-y-auto` is the fallback page scroller for subtrees that don't opt into the chain.
        */}
        <div className="pt-[3.25rem] flex-1 min-h-0 flex flex-col overflow-y-auto">
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
