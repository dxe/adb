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
import { redirectIfForbidden } from '@/lib/server-auth'
import ActivistsPage from './activists-page'
import { buildQueryOptions } from './filter-api-query'
import {
  getActivistQueryStateFromParams,
  loadActivistSearchParams,
} from './search-params'

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
  const { filters, selectedColumns, sort } =
    getActivistQueryStateFromParams(parsedSearchParams)

  const queryClient = new QueryClient()
  const apiClient = new ApiClient(cookies)
  const initialReferenceDate = new Date()

  const initialQueryOptions = buildQueryOptions({
    filters,
    selectedColumns,
    chapterId: session.user.ChapterID,
    userId: session.user.ID,
    referenceDate: initialReferenceDate,
    sort,
  })

  await redirectIfForbidden(() =>
    // Use fetchInfiniteQuery instead of prefetchInfiniteQuery so a 403 throws
    // during SSR and redirectIfForbidden can trigger Next's forbidden UI immediately.
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
            initialReferenceDateIso={initialReferenceDate.toISOString()}
          />
        </Suspense>
      </HydrationBoundary>
    </ContentWrapper>
  )
}
