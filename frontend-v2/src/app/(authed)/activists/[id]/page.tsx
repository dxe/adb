import {
  dehydrate,
  HydrationBoundary,
  QueryClient,
} from '@tanstack/react-query'
import { notFound } from 'next/navigation'
import Link from 'next/link'
import { ArrowLeft } from 'lucide-react'
import { ContentWrapper } from '@/app/content-wrapper'
import { API_PATH, ApiClient } from '@/lib/api'
import { getCookies } from '@/lib/auth'
import { parseSafeInteger } from '@/lib/number-utils'
import { redirectForHttpError } from '@/lib/server-auth'
import { ActivistDetail } from './activist-detail'

export default async function ActivistPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = await params
  const activistId = parseSafeInteger(id)
  if (activistId === undefined) {
    notFound()
  }

  const apiClient = new ApiClient(await getCookies())
  const queryClient = new QueryClient()

  await redirectForHttpError(() =>
    // Intentionally use fetchQuery instead of prefetchQuery; see redirectForHttpError for details.
    queryClient.fetchQuery({
      queryKey: [API_PATH.ACTIVIST_GET, activistId],
      queryFn: ({ signal }) => apiClient.getActivist(activistId, signal),
    }),
  )

  return (
    <ContentWrapper size="lg" className="gap-6">
      <div className="flex items-center gap-3">
        <Link
          href="/activists"
          className="flex items-center gap-1 text-sm text-muted-foreground transition-colors hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" />
          View all Activists
        </Link>
      </div>
      <HydrationBoundary state={dehydrate(queryClient)}>
        <ActivistDetail activistId={activistId} />
      </HydrationBoundary>
    </ContentWrapper>
  )
}
