import 'styles/globals.css'
import type { Metadata, Viewport } from 'next'
import { Toaster } from 'react-hot-toast'
import Providers from './providers'

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
      {/*
      Without "has-navbar-fixed-top" we get an error:
      "Hydration failed because the server rendered HTML didn't match the
      client. As a result this tree will be regenerated on the client...."
      */}
      <body className="antialiased has-navbar-fixed-top">
        <Providers>{children}</Providers>
        <Toaster position="bottom-right" />
      </body>
    </html>
  )
}
