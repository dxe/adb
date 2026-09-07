import { Suspense } from 'react'
import {
  dehydrate,
  HydrationBoundary,
  QueryClient,
} from '@tanstack/react-query'
import { notFound } from 'next/navigation'
import { ContentWrapper } from '@/app/content-wrapper'
import { ApiClient, CHAPTER_ADMIN_QUERY_KEY } from '@/lib/api'
import { getCookies } from '@/lib/auth'
import { redirectForHttpError } from '@/lib/server-auth'
import { FormLoadingFallback } from '../form-loading-fallback'
import { ChapterForm } from '../chapter-form'

export default async function EditChapterPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = await params
  if (!/^\d+$/.test(id)) {
    notFound()
  }
  const chapterId = Number(id)

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
    <ContentWrapper size="lg" className="gap-6">
      <HydrationBoundary state={dehydrate(queryClient)}>
        <Suspense fallback={<FormLoadingFallback />}>
          <ChapterForm chapterId={chapterId} />
        </Suspense>
      </HydrationBoundary>
    </ContentWrapper>
  )
}
