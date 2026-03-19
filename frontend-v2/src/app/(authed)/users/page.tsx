import {
  dehydrate,
  HydrationBoundary,
  QueryClient,
} from '@tanstack/react-query'
import { ContentWrapper } from '@/app/content-wrapper'
import { API_PATH, ApiClient } from '@/lib/api'
import { getCookies } from '@/lib/auth'
import { redirectForHttpError } from '@/lib/server-auth'
import UsersPage from './users-page'

export default async function UsersListPage() {
  const apiClient = new ApiClient(await getCookies())
  const queryClient = new QueryClient()

  await redirectForHttpError(() =>
    Promise.all([
      // Intentionally use fetchQuery instead of prefetchQuery; see redirectForHttpError for details.
      queryClient.fetchQuery({
        queryKey: [API_PATH.USERS],
        queryFn: ({ signal }) => apiClient.getUsers(signal),
      }),
      // Intentionally use fetchQuery instead of prefetchQuery; see redirectForHttpError for details.
      queryClient.fetchQuery({
        queryKey: [API_PATH.CHAPTER_LIST],
        queryFn: ({ signal }) => apiClient.getChapterList(signal),
      }),
    ]),
  )

  return (
    <ContentWrapper size="xl" className="gap-6">
      <HydrationBoundary state={dehydrate(queryClient)}>
        <UsersPage />
      </HydrationBoundary>
    </ContentWrapper>
  )
}
