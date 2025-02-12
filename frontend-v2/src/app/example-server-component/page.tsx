// Modeled from example from React Query docs:
// https://tanstack.com/query/latest/docs/framework/react/guides/advanced-ssr

import {
  dehydrate,
  HydrationBoundary,
  QueryClient,
} from '@tanstack/react-query'
import Activists from './activists'
import { VueNavbar } from 'app/VueNavbar'
import { ContentWrapper } from 'app/ContentWrapper'
import { AuthedPageLayout } from 'app/AuthedPageLayout'
import { API_PATH, ApiClient } from 'lib/api'
import { getAuthCookies } from 'lib/auth'

export default async function ActivistsPage() {
  const apiClient = new ApiClient(await getAuthCookies())
  const queryClient = new QueryClient()

  await queryClient.prefetchQuery({
    queryKey: [API_PATH.ACTIVIST_NAMES_GET],
    queryFn: apiClient.getActivistNames,
  })

  return (
    <AuthedPageLayout>
      <VueNavbar pageName="TestPage" />
      <ContentWrapper size="sm" className="gap-6">
        <p>Hello from App Router!</p>

        {
          // Serialization is as easy as passing props.
          // HydrationBoundary is a Client Component, so hydration will happen there.
        }
        <HydrationBoundary state={dehydrate(queryClient)}>
          <Activists />
        </HydrationBoundary>
      </ContentWrapper>
    </AuthedPageLayout>
  )
}
