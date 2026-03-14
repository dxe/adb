import {
  dehydrate,
  HydrationBoundary,
  QueryClient,
} from '@tanstack/react-query'
import { ContentWrapper } from '@/app/content-wrapper'
import { API_PATH, ApiClient } from '@/lib/api'
import { getCookies } from '@/lib/auth'
import { redirectIfForbidden } from '@/lib/server-auth'
import UsersPage from './users-page'

export default async function UsersListPage() {
  const apiClient = new ApiClient(await getCookies())
  const queryClient = new QueryClient()

  // Use fetchQuery instead of prefetchQuery so a 403 throws during SSR
  // and redirectIfForbidden can send the user to /403 immediately.
  await redirectIfForbidden(() =>
    Promise.all([
      queryClient.fetchQuery({
        queryKey: [API_PATH.USERS],
        queryFn: ({ signal }) => apiClient.getUsers(signal),
      }),
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
