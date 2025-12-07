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
import { UserForm } from '../user-form'

export default async function EditUserPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const userId = Number((await params).id)
  const apiClient = new ApiClient(await getCookies())
  const queryClient = new QueryClient()

  await Promise.all([
    queryClient.prefetchQuery({
      queryKey: [API_PATH.USERS, userId],
      queryFn: () => apiClient.getUser(userId),
    }),
    queryClient.prefetchQuery({
      queryKey: [API_PATH.CHAPTER_LIST],
      queryFn: apiClient.getChapterList,
    }),
  ])

  return (
    <AuthedPageLayout pageName="UserList">
      <HydrationBoundary state={dehydrate(queryClient)}>
        <Navbar />
        <ContentWrapper size="lg" className="gap-6">
          <UserForm userId={userId} />
        </ContentWrapper>
      </HydrationBoundary>
    </AuthedPageLayout>
  )
}
