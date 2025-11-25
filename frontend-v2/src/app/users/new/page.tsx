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

export default async function NewUserPage() {
  const apiClient = new ApiClient(await getCookies())
  const queryClient = new QueryClient()

  await queryClient.prefetchQuery({
    queryKey: [API_PATH.CHAPTER_LIST],
    queryFn: apiClient.getChapterList,
  })

  return (
    <AuthedPageLayout pageName="UserList">
      <Navbar />
      <ContentWrapper size="lg" className="gap-6">
        <HydrationBoundary state={dehydrate(queryClient)}>
          <UserForm />
        </HydrationBoundary>
      </ContentWrapper>
    </AuthedPageLayout>
  )
}
