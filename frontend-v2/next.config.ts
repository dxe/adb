import type { NextConfig } from 'next'

const nextConfig: NextConfig = {
  /* config options here */
  reactStrictMode: true,
  images: {
    unoptimized: true,
  },
  basePath: '/v2',
  output: 'standalone',
}

export default nextConfig
