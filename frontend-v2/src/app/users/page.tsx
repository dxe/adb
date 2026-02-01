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
import UsersPage from './users-page'

export default async function UsersListPage() {
  const apiClient = new ApiClient(await getCookies())
  const queryClient = new QueryClient()

  // Prefetch data on server
  await Promise.all([
    queryClient.prefetchQuery({
      queryKey: [API_PATH.USERS],
      queryFn: apiClient.getUsers,
    }),
    queryClient.prefetchQuery({
      queryKey: [API_PATH.CHAPTER_LIST],
      queryFn: apiClient.getChapterList,
    }),
  ])

  return (
    <AuthedPageLayout pageName="UserList">
      <Navbar />
      <ContentWrapper size="xl" className="gap-6">
        <HydrationBoundary state={dehydrate(queryClient)}>
          <UsersPage />
        </HydrationBoundary>
      </ContentWrapper>
    </AuthedPageLayout>
  )
}
