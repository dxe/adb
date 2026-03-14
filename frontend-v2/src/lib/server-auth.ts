import 'server-only'
import { forbidden } from 'next/navigation'
import { HTTPStatusError } from '@/lib/api'

export async function redirectIfForbidden<T>(
  load: () => Promise<T>,
): Promise<T> {
  try {
    return await load()
  } catch (err) {
    if (err instanceof HTTPStatusError && err.status === 403) {
      forbidden()
    }
    throw err
  }
}
