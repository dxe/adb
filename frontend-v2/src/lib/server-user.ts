import { headers } from 'next/headers'
import { User } from '@/app/session'

export const SERVER_USER_HEADER = 'x-user-data'

/**
 * Get the authenticated user from request headers (set by middleware).
 * Only works in server components.
 */
export async function getServerUser(): Promise<User> {
  const headersList = await headers()
  const userHeader = headersList.get(SERVER_USER_HEADER)

  if (!userHeader) {
    throw new Error(
      'User not found in headers. This should not happen if middleware is working correctly.',
    )
  }

  return JSON.parse(userHeader) as User
}
