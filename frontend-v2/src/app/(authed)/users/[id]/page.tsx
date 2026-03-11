import {
  dehydrate,
  HydrationBoundary,
  QueryClient,
} from '@tanstack/react-query'
import { notFound } from 'next/navigation'
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
  const userId = parseInt(id)
  if (Number.isNaN(userId)) {
    notFound()
  }

  const apiClient = new ApiClient(await getCookies())
  const queryClient = new QueryClient()

  await Promise.all([
    queryClient.prefetchQuery({
      queryKey: [API_PATH.USERS, userId],
      queryFn: ({ signal }) => apiClient.getUser(userId, signal),
    }),
    queryClient.prefetchQuery({
      queryKey: [API_PATH.CHAPTER_LIST],
      queryFn: ({ signal }) => apiClient.getChapterList(signal),
    }),
  ])

  return (
    <ContentWrapper size="lg" className="gap-6">
      <HydrationBoundary state={dehydrate(queryClient)}>
        <UserForm userId={userId} />
      </HydrationBoundary>
    </ContentWrapper>
  )
}
