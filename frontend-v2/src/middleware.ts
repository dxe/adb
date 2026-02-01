import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'
import { fetchSession, type User } from '@/app/session'
import { SERVER_USER_HEADER } from '@/lib/server-user'

// Simple in-memory cache for middleware (Edge Runtime compatible)
// Maps cookie hash to { user, timestamp }
const sessionCache = new Map<string, { user: User; timestamp: number }>()
const CACHE_TTL = 3600 * 1000 // 1 hour in milliseconds
const MAX_CACHE_SIZE = 1000 // Maximum number of cached sessions before cleanup

// Simple hash function for cookie string (for cache key)
function hashString(str: string): string {
  let hash = 0
  for (let i = 0; i < str.length; i++) {
    const char = str.charCodeAt(i)
    hash = (hash << 5) - hash + char
    hash = hash & hash // Convert to 32-bit integer
  }
  return hash.toString(36)
}

async function getCachedSession(cookies: string) {
  const cacheKey = hashString(cookies)
  const cached = sessionCache.get(cacheKey)
  const now = Date.now()

  // Return cached if valid
  if (cached && now - cached.timestamp < CACHE_TTL) {
    return { user: cached.user }
  }

  // Fetch fresh session
  const session = await fetchSession(cookies)

  // Cache it
  if (session.user) {
    sessionCache.set(cacheKey, {
      user: session.user,
      timestamp: now,
    })

    // Cleanup expired entries on every cache set
    const cutoff = now - CACHE_TTL
    for (const [key, value] of sessionCache.entries()) {
      if (value.timestamp < cutoff) {
        sessionCache.delete(key)
      }
    }

    // Additional cleanup if cache grows too large (safety net)
    if (sessionCache.size > MAX_CACHE_SIZE) {
      // Remove oldest entries first
      const entries = Array.from(sessionCache.entries())
      entries.sort((a, b) => a[1].timestamp - b[1].timestamp)
      const toRemove = sessionCache.size - MAX_CACHE_SIZE
      for (let i = 0; i < toRemove; i++) {
        sessionCache.delete(entries[i][0])
      }
    }
  }

  return session
}

export async function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl

  // Allow public routes
  if (pathname.startsWith('/login') || pathname.startsWith('/auth')) {
    return NextResponse.next()
  }

  // Get cookies from request
  const cookieHeader = request.headers.get('cookie') || ''

  // Fetch session (cached)
  const session = await getCachedSession(cookieHeader)

  // Redirect to login if no user
  if (!session.user) {
    const loginUrl = new URL('/login', request.url)
    return NextResponse.redirect(loginUrl)
  }

  // Attach user data to headers so server components can read it
  const requestHeaders = new Headers(request.headers)
  requestHeaders.set(SERVER_USER_HEADER, JSON.stringify(session.user))

  // Continue with modified headers
  return NextResponse.next({
    request: {
      headers: requestHeaders,
    },
  })
}

// Configure which routes use this middleware
export const config = {
  matcher: [
    /*
     * Match all request paths except for:
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     * - public files (public folder)
     */
    '/((?!_next/static|_next/image|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp)$).*)',
  ],
}
