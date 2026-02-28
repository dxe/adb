import { headers } from 'next/headers'
import { redirect } from 'next/navigation'
import { User } from '@/app/session'

export const SERVER_USER_HEADER = 'x-user-data'

/**
 * Get the authenticated user from request headers (set by middleware).
 * Only works in server components.
 * Redirects to /login if user header is missing or malformed.
 */
export async function getServerUser(): Promise<User> {
  const headersList = await headers()
  const userHeader = headersList.get(SERVER_USER_HEADER)

  if (!userHeader) {
    console.error('User header missing - redirecting to login')
    redirect('/login')
  }

  try {
    return JSON.parse(userHeader) as User
  } catch {
    console.error('User header malformed - redirecting to login')
    redirect('/login')
  }
}
