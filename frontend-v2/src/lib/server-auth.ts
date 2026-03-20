import 'server-only'
import { forbidden, notFound } from 'next/navigation'
import { HTTPStatusError } from '@/lib/api'

// Wrap SSR data loading that should surface HTTP errors as Next route UI.
// Call sites must intentionally use fetchQuery/fetchInfiniteQuery instead of
// prefetchQuery/prefetchInfiniteQuery; the prefetch variants swallow errors, so
// 403/404 responses would bypass this helper and silently skip Next's
// forbidden/not-found handling.
export async function redirectForHttpError<T>(
  load: () => Promise<T>,
): Promise<T> {
  try {
    return await load()
  } catch (err) {
    if (err instanceof HTTPStatusError) {
      if (err.status === 403) {
        forbidden()
      }
      if (err.status === 404) {
        notFound()
      }
    }
    throw err
  }
}
