import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'
import { fetchSession, type User } from '@/app/session'
import { SERVER_USER_HEADER } from '@/lib/server-user'
import QuickLRU from 'quick-lru'

const CACHE_TTL = 15 * 60 * 1000 // 15 minutes in milliseconds

// LRU cache for middleware (Edge Runtime compatible)
// Maps cookie hash to user; QuickLRU handles TTL-based expiry automatically.
const sessionCache = new QuickLRU<string, User>({
  maxSize: 1000,
  maxAge: CACHE_TTL,
})

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

  if (cached) return { user: cached }

  const session = await fetchSession(cookies)

  if (session.user) {
    sessionCache.set(cacheKey, session.user)
  }

  return session
}

export async function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl

  // Allow public routes
  if (pathname.startsWith('/login') || pathname.startsWith('/auth')) {
    return NextResponse.next()
  }

  const cookieHeader = request.headers.get('cookie') || ''

  let session
  try {
    session = await getCachedSession(cookieHeader)
  } catch (error) {
    console.error('Session fetch failed - redirecting to login:', error)
    const loginUrl = new URL('/login', request.url)
    return NextResponse.redirect(loginUrl)
  }

  if (!session.user) {
    const loginUrl = new URL('/login', request.url)
    return NextResponse.redirect(loginUrl)
  }

  // Attach user data to headers so server components can read it
  const requestHeaders = new Headers(request.headers)
  requestHeaders.set(SERVER_USER_HEADER, JSON.stringify(session.user))

  return NextResponse.next({
    request: {
      headers: requestHeaders,
    },
  })
}

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
