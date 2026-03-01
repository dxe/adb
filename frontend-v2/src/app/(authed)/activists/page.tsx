import {
  dehydrate,
  HydrationBoundary,
  QueryClient,
} from '@tanstack/react-query'
import { ContentWrapper } from '@/app/content-wrapper'
import { API_PATH, ApiClient } from '@/lib/api'
import { getCookies } from '@/lib/auth'
import { getCachedSession } from '@/app/session'
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
  const [cookies, session] = await Promise.all([
    getCookies(),
    getCachedSession(), // deduplicated with layout — no extra network call
  ])

  if (!session.user) {
    throw new Error('missing user in session')
  }

  const params = await searchParams
  const getParam = (key: string) => {
    const value = params[key]
    if (Array.isArray(value)) return value[0]
    return value
  }

  const filters = parseFiltersFromParams(getParam)
  const selectedColumns = parseColumnsFromParams(getParam)
  const normalizedColumns = normalizeColumnsForFilters(
    selectedColumns,
    filters.searchAcrossChapters,
  )
  const sort = parseSortFromParams(getParam, normalizedColumns)

  const queryClient = new QueryClient()
  const apiClient = new ApiClient(cookies)

  const initialQueryOptions = buildQueryOptions({
    filters,
    selectedColumns,
    chapterId: session.user.ChapterID,
    sort,
  })

  await queryClient.prefetchInfiniteQuery({
    queryKey: [API_PATH.ACTIVISTS_SEARCH, initialQueryOptions],
    queryFn: () => apiClient.searchActivists(initialQueryOptions),
    initialPageParam: undefined as string | undefined,
  })

  return (
    <ContentWrapper size="xl" className="gap-6">
      <HydrationBoundary state={dehydrate(queryClient)}>
        <ActivistsPage />
      </HydrationBoundary>
    </ContentWrapper>
  )
}
