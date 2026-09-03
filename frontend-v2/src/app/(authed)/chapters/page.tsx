import {
  dehydrate,
  HydrationBoundary,
  QueryClient,
} from '@tanstack/react-query'
import { ContentWrapper } from '@/app/content-wrapper'
import { ApiClient, CHAPTER_ADMIN_QUERY_KEY } from '@/lib/api'
import { getCookies } from '@/lib/auth'
import { redirectForHttpError } from '@/lib/server-auth'
import ChaptersPage from './chapters-page'

export default async function ChaptersListPage() {
  const apiClient = new ApiClient(await getCookies())
  const queryClient = new QueryClient()

  await redirectForHttpError(() =>
    // Intentionally use fetchQuery instead of prefetchQuery; see redirectForHttpError for details.
    queryClient.fetchQuery({
      queryKey: CHAPTER_ADMIN_QUERY_KEY,
      queryFn: ({ signal }) => apiClient.getChapterAdminList(signal),
    }),
  )

  return (
    <ContentWrapper size="2xl" className="gap-6">
      <HydrationBoundary state={dehydrate(queryClient)}>
        <ChaptersPage />
      </HydrationBoundary>
    </ContentWrapper>
  )
}
