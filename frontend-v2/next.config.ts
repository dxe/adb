import path from 'path'
import type { NextConfig } from 'next'

const nextConfig: NextConfig = {
  /* config options here */
  reactStrictMode: true,
  images: {
    unoptimized: true,
  },
  basePath: '/v2',
  output: 'standalone',
  turbopack: {
    // Ensure the repo root includes the ../shared directory. Required in the Dockerfile environment. Locally,
    // Turbopack may detect the workspace's pnpm-lock.yaml and thus decide to use the workspace as the root anyway.
    root: path.resolve(__dirname, '..'),
  },
}

export default nextConfig
