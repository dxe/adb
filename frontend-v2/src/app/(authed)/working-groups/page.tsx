import {
  dehydrate,
  HydrationBoundary,
  QueryClient,
} from '@tanstack/react-query'
import { ContentWrapper } from '@/app/content-wrapper'
import { API_PATH, ApiClient } from '@/lib/api'
import { getCookies } from '@/lib/auth'
import { redirectForHttpError } from '@/lib/server-auth'
import WorkingGroupsPage from './working-groups-page'

// Gating for this page happens server-side: the working_group/* endpoints
// require SF Bay organizer access (admin, or organizer + SF Bay chapter). A
// non-authorized user's fetchQuery below throws HTTPStatusError(403), which
// redirectForHttpError turns into Next's forbidden() page. See users/page.tsx
// for the same pattern applied to admin-only endpoints.
export default async function WorkingGroupsListPage() {
  const apiClient = new ApiClient(await getCookies())
  const queryClient = new QueryClient()

  await redirectForHttpError(() =>
    Promise.all([
      // Intentionally use fetchQuery instead of prefetchQuery; see redirectForHttpError for details.
      queryClient.fetchQuery({
        queryKey: [API_PATH.WORKING_GROUP_LIST],
        queryFn: ({ signal }) => apiClient.getWorkingGroups(signal),
      }),
      queryClient.fetchQuery({
        queryKey: [API_PATH.ACTIVIST_NAMES_GET_ORGANIZERS],
        queryFn: ({ signal }) => apiClient.getOrganizerNames(signal),
      }),
      queryClient.fetchQuery({
        queryKey: [API_PATH.ACTIVIST_NAMES_GET],
        queryFn: ({ signal }) => apiClient.getActivistNames(signal),
      }),
    ]),
  )

  return (
    <ContentWrapper size="xl" className="gap-6">
      <HydrationBoundary state={dehydrate(queryClient)}>
        <WorkingGroupsPage />
      </HydrationBoundary>
    </ContentWrapper>
  )
}
