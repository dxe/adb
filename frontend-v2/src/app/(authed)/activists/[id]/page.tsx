import {
  dehydrate,
  HydrationBoundary,
  QueryClient,
} from '@tanstack/react-query'
import { notFound } from 'next/navigation'
import { ContentWrapper } from '@/app/content-wrapper'
import { API_PATH, ApiClient } from '@/lib/api'
import { getCookies } from '@/lib/auth'
import { redirectIfForbidden } from '@/lib/server-auth'
import { ActivistDetail } from './activist-detail'

export default async function ActivistPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = await params
  const activistId = parseInt(id)
  if (Number.isNaN(activistId)) {
    notFound()
  }

  const apiClient = new ApiClient(await getCookies())
  const queryClient = new QueryClient()

  // Use fetchQuery instead of prefetchQuery so a 403 throws during SSR
  // and redirectIfForbidden can trigger Next's forbidden UI immediately.
  await redirectIfForbidden(() =>
    queryClient.fetchQuery({
      queryKey: [API_PATH.ACTIVIST_GET, activistId],
      queryFn: ({ signal }) => apiClient.getActivist(activistId, signal),
    }),
  )

  return (
    <ContentWrapper size="lg" className="gap-6">
      <HydrationBoundary state={dehydrate(queryClient)}>
        <ActivistDetail activistId={activistId} />
      </HydrationBoundary>
    </ContentWrapper>
  )
}
