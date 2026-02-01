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
  const { id } = await params
  const userId = Number(id)
  const apiClient = new ApiClient(await getCookies())
  const queryClient = new QueryClient()

  // Prefetch user data on server
  await queryClient.prefetchQuery({
    queryKey: [API_PATH.USERS, userId],
    queryFn: () => apiClient.getUser(userId),
  })

  return (
    <AuthedPageLayout pageName="UserList">
      <Navbar />
      <ContentWrapper size="lg" className="gap-6">
        <HydrationBoundary state={dehydrate(queryClient)}>
          <UserForm userId={userId} />
        </HydrationBoundary>
      </ContentWrapper>
    </AuthedPageLayout>
  )
}
