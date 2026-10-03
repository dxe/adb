import path from 'path'
import type { NextConfig } from 'next'

const nextConfig: NextConfig = {
  /* config options here */
  reactStrictMode: true,
  experimental: {
    authInterrupts: true,
  },
  images: {
    unoptimized: true,
  },
  basePath: '/v2',
  output: 'standalone',
  turbopack: {
    // Keep the repo root: Dockerfile.frontend-v2 relies on the standalone
    // output nesting under .next/standalone/frontend-v2/.
    root: path.resolve(__dirname, '..'),
  },
  allowedDevOrigins: ['[::1]'],
}

export default nextConfig
