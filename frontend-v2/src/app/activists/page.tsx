import {
  dehydrate,
  HydrationBoundary,
  QueryClient,
} from '@tanstack/react-query'
import { AuthedPageLayout } from '@/app/authed-page-layout'
import { Navbar } from '@/components/nav'
import { ContentWrapper } from '@/app/content-wrapper'
import { API_PATH, ApiClient } from '@/lib/api'
import { getCookies } from '@/lib/auth'
import { fetchSession } from '@/app/session'
import ActivistsPage from './activists-page'
import {
  buildQueryOptions,
  parseColumnsFromParams,
  parseFiltersFromParams,
  parseSortFromParams,
} from './query-utils'
import { normalizeColumnsForFilters } from './column-definitions'

interface PageProps {
  searchParams: Promise<{ [key: string]: string | string[] | undefined }>
}

export default async function ActivistsListPage({ searchParams }: PageProps) {
  const cookies = await getCookies()
  const apiClient = new ApiClient(cookies)
  const queryClient = new QueryClient()
  const session = await fetchSession(cookies)

  if (!session.user) {
    throw new Error('missing user in session')
  }

  // Parse filter options from URL
  const params = await searchParams
  const getParam = (key: string) => {
    const value = params[key]
    if (Array.isArray(value)) {
      return value[0]
    }
    return value
  }
  const filters = parseFiltersFromParams(getParam)
  const selectedColumns = parseColumnsFromParams(getParam)
  const normalizedColumns = normalizeColumnsForFilters(
    selectedColumns,
    filters.searchAcrossChapters,
  )
  const sort = parseSortFromParams(getParam, normalizedColumns)
  const initialQueryOptions = buildQueryOptions({
    filters,
    selectedColumns,
    chapterId: session.user.ChapterID,
    userId: session.user.ID,
    sort,
  })

  // Prefetch first page of activists data for SSR hydration
  await queryClient.prefetchInfiniteQuery({
    queryKey: [API_PATH.ACTIVISTS_SEARCH, initialQueryOptions],
    queryFn: () => apiClient.searchActivists(initialQueryOptions),
    initialPageParam: undefined as string | undefined,
  })

  return (
    <AuthedPageLayout pageName="ActivistList">
      <HydrationBoundary state={dehydrate(queryClient)}>
        <Navbar />
        <ContentWrapper size="xl" className="gap-6">
          <ActivistsPage />
        </ContentWrapper>
      </HydrationBoundary>
    </AuthedPageLayout>
  )
}
