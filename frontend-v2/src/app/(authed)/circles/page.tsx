import {
  dehydrate,
  HydrationBoundary,
  QueryClient,
} from '@tanstack/react-query'
import { ContentWrapper } from '@/app/content-wrapper'
import { API_PATH, ApiClient } from '@/lib/api'
import { getCookies } from '@/lib/auth'
import { redirectForHttpError } from '@/lib/server-auth'
import CirclesPage from './circles-page'

// Circles ("Interest Circles" and "Geo-Circles" in the nav) are gated by the
// Go backend's `apiSFBayOrganizerAuthMiddleware` on the endpoints prefetched
// below (admins always have access; organizers only if their chapter is SF
// Bay). `redirectForHttpError` surfaces the backend's 403 as Next's forbidden
// UI, so no role logic is duplicated here.
export default async function CircleGroupsPage() {
  const apiClient = new ApiClient(await getCookies())
  const queryClient = new QueryClient()

  await redirectForHttpError(() =>
    Promise.all([
      // Intentionally use fetchQuery instead of prefetchQuery; see redirectForHttpError for details.
      queryClient.fetchQuery({
        queryKey: [API_PATH.CIRCLE_LIST],
        queryFn: ({ signal }) => apiClient.getCircles(signal),
      }),
      queryClient.fetchQuery({
        queryKey: [API_PATH.ACTIVIST_NAMES_CHAPTER_MEMBERS],
        queryFn: ({ signal }) =>
          apiClient.getChapterMemberActivistNames(signal),
      }),
    ]),
  )

  return (
    <ContentWrapper size="xl" className="gap-6">
      <HydrationBoundary state={dehydrate(queryClient)}>
        <CirclesPage />
      </HydrationBoundary>
    </ContentWrapper>
  )
}
