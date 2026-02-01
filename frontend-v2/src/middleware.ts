import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'
import { fetchSession } from '@/app/session'
import { SERVER_USER_HEADER } from '@/lib/server-user'

export async function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl

  // Allow public routes
  if (pathname.startsWith('/login') || pathname.startsWith('/auth')) {
    return NextResponse.next()
  }

  // Get cookies from request
  const cookieHeader = request.headers.get('cookie') || ''

  // Fetch session
  const session = await fetchSession(cookieHeader)

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
