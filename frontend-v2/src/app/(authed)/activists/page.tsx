import {
  dehydrate,
  HydrationBoundary,
  QueryClient,
} from '@tanstack/react-query'
import { Suspense } from 'react'
import { ContentWrapper } from '@/app/content-wrapper'
import { API_PATH, ApiClient } from '@/lib/api'
import { getCookies } from '@/lib/auth'
import { getCachedSession } from '@/app/session'
import { redirectForHttpError } from '@/lib/server-auth'
import ActivistsPage from './activists-page'
import { buildQueryOptions } from './filter-api-query'
import {
  getActivistQueryStateFromParams,
  loadActivistSearchParams,
} from './search-params'
import type { ActivistsQueryState } from './query-state'

interface PageProps {
  searchParams: Promise<{ [key: string]: string | string[] | undefined }>
}

export default async function ActivistsListPage({ searchParams }: PageProps) {
  const [cookies, session] = await Promise.all([
    getCookies(),
    getCachedSession(), // de-duped since this is also called in layout
  ])

  if (!session.user) {
    throw new Error('missing user in session')
  }

  const parsedSearchParams = await loadActivistSearchParams(searchParams)
  const initialParamQueryState =
    getActivistQueryStateFromParams(parsedSearchParams)

  const queryClient = new QueryClient()
  const apiClient = new ApiClient(cookies)
  const initialReferenceDate = new Date()

  const debugInitialServerQueryState: ActivistsQueryState | undefined =
    process.env.NODE_ENV === 'development' ? initialParamQueryState : undefined

  const initialQueryOptions = buildQueryOptions({
    ...initialParamQueryState,
    chapterId: session.user.ChapterID,
    userId: session.user.ID,
    referenceDate: initialReferenceDate,
  })

  await redirectForHttpError(() =>
    // Intentionally use fetchInfiniteQuery instead of prefetchInfiniteQuery; see redirectForHttpError for details.
    queryClient.fetchInfiniteQuery({
      queryKey: [API_PATH.ACTIVISTS_SEARCH, initialQueryOptions],
      queryFn: ({ signal }) =>
        apiClient.searchActivists(initialQueryOptions, signal),
      initialPageParam: undefined as string | undefined,
    }),
  )

  return (
    <ContentWrapper size="full" className="gap-6">
      <HydrationBoundary state={dehydrate(queryClient)}>
        <Suspense fallback={null}>
          <ActivistsPage
            debugInitialServerQueryState={debugInitialServerQueryState}
            initialReferenceDateIso={initialReferenceDate.toISOString()}
          />
        </Suspense>
      </HydrationBoundary>
    </ContentWrapper>
  )
}
