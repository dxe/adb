import { forbidden } from 'next/navigation'
import {
  dehydrate,
  HydrationBoundary,
  QueryClient,
} from '@tanstack/react-query'
import { evaluateNavAccess } from '$shared/nav-access'
import { ContentWrapper } from '@/app/content-wrapper'
import { getCachedSession } from '@/app/session'
import { API_PATH, ApiClient } from '@/lib/api'
import { getCookies } from '@/lib/auth'
import { SF_BAY_CHAPTER_ID } from '@/lib/constants'
import { redirectForHttpError } from '@/lib/server-auth'
import CirclesPage from './circles-page'

// Circles ("Interest Circles" and "Geo-Circles" in the nav) are gated
// server-side by `authSFBayOrganizerMiddleware`/`apiSFBayOrganizerAuthMiddleware`
// in Go: admins always have access; organizers only if their chapter is SF Bay.
const CIRCLES_NAV_ACCESS_RULE = {
  roleRequired: ['organizer'],
  visibleForNonSFBay: false,
}

export default async function CircleGroupsPage() {
  const session = await getCachedSession()

  if (
    !session.user ||
    !evaluateNavAccess(
      session.user.Roles,
      session.user.ChapterID,
      CIRCLES_NAV_ACCESS_RULE,
      SF_BAY_CHAPTER_ID,
    )
  ) {
    forbidden()
  }

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
