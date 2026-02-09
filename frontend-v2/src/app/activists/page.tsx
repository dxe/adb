import {
  dehydrate,
  HydrationBoundary,
  QueryClient,
} from '@tanstack/react-query'
import { AuthedPageLayout } from '@/app/authed-page-layout'
import { Navbar } from '@/components/nav'
import { ContentWrapper } from '@/app/content-wrapper'
import {
  API_PATH,
  ApiClient,
  QueryActivistOptions,
  ActivistColumnName,
} from '@/lib/api'
import { getCookies } from '@/lib/auth'
import { fetchSession } from '@/app/session'
import ActivistsPage from './activists-page'
import { getDefaultColumns, sortColumnsByDefinitionOrder } from './column-definitions'

interface PageProps {
  searchParams: Promise<{ [key: string]: string | string[] | undefined }>
}

export default async function ActivistsListPage({ searchParams }: PageProps) {
  const cookies = await getCookies()
  const apiClient = new ApiClient(cookies)
  const queryClient = new QueryClient()
  const session = await fetchSession(cookies)

  if (!session.user) {
    // This shouldn't happen due to AuthedPageLayout, but handle gracefully
    return null
  }

  // Parse URL search params
  const params = await searchParams
  const showAllChapters = params.showAllChapters === 'true'
  const nameSearch = typeof params.nameSearch === 'string' ? params.nameSearch : ''
  const lastEventGte =
    typeof params.lastEventGte === 'string' ? params.lastEventGte : undefined
  const lastEventLt =
    typeof params.lastEventLt === 'string' ? params.lastEventLt : undefined
  const columnsParam = typeof params.columns === 'string' ? params.columns : ''

  // Parse columns from URL or use defaults
  const defaultColumns = getDefaultColumns(false)
  const parsedColumns: ActivistColumnName[] = columnsParam
    ? sortColumnsByDefinitionOrder(columnsParam.split(',') as ActivistColumnName[])
    : defaultColumns

  // Build columns list for API request
  let columnsToRequest = [...parsedColumns]

  // Add chapter_name if showing all chapters
  if (showAllChapters && !columnsToRequest.includes('chapter_name')) {
    columnsToRequest.unshift('chapter_name')
  }

  // Always include ID for row keys
  if (!columnsToRequest.includes('id')) {
    columnsToRequest.unshift('id')
  }

  // Build initial query options
  const initialQueryOptions: QueryActivistOptions = {
    columns: columnsToRequest,
    filters: {
      chapter_id: showAllChapters ? 0 : session.user.ChapterID,
      name: nameSearch ? { name_contains: nameSearch } : undefined,
      last_event:
        lastEventGte || lastEventLt
          ? {
              last_event_gte: lastEventGte,
              last_event_lt: lastEventLt,
            }
          : undefined,
    },
  }

  // Prefetch initial activists data
  await queryClient.prefetchQuery({
    queryKey: [API_PATH.ACTIVISTS_SEARCH, initialQueryOptions],
    queryFn: () => apiClient.searchActivists(initialQueryOptions),
  })

  // Prefetch chapter list (may be needed for future features)
  await queryClient.prefetchQuery({
    queryKey: [API_PATH.CHAPTER_LIST],
    queryFn: apiClient.getChapterList,
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
