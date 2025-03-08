import type { Metadata, Viewport } from 'next'
import { Toaster } from 'react-hot-toast'
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
      {/* Top padding is to make room for the fixed navbar. */}
      <body className="antialiased pt-[3.25rem]">
        <Providers>{children}</Providers>
        <Toaster position="bottom-right" />
      </body>
    </html>
  )
}
