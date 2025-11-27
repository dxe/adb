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

  await queryClient.prefetchQuery({
    queryKey: [API_PATH.USERS],
    queryFn: apiClient.getUsers,
  })

  await queryClient.prefetchQuery({
    queryKey: [API_PATH.CHAPTER_LIST],
    queryFn: apiClient.getChapterList,
  })

  return (
    <AuthedPageLayout pageName="UserList">
      <HydrationBoundary state={dehydrate(queryClient)}>
        <Navbar />
        <ContentWrapper size="xl" className="gap-6">
          <UsersPage />
        </ContentWrapper>
      </HydrationBoundary>
    </AuthedPageLayout>
  )
}
