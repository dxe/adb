// Modeled from example from React Query docs:
// https://tanstack.com/query/latest/docs/framework/react/guides/advanced-ssr

import {
  dehydrate,
  HydrationBoundary,
  QueryClient,
} from '@tanstack/react-query'
import Activists from './activists'
import { ContentWrapper } from '@/app/content-wrapper'
import { AuthedPageLayout } from '@/app/authed-page-layout'
import { API_PATH, ApiClient } from '@/lib/api'
import { getCookies } from '@/lib/auth'
import { fetchSession } from '../session'
import { Navbar } from '@/components/nav'

export default async function ActivistsPage() {
  const apiClient = new ApiClient(await getCookies())
  const queryClient = new QueryClient()
  const session = await fetchSession(await getCookies())

  await queryClient.prefetchQuery({
    queryKey: [API_PATH.ACTIVIST_NAMES_GET],
    queryFn: apiClient.getActivistNames,
  })

  return (
    <AuthedPageLayout>
      <Navbar pageName="TestPage" session={session} />
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
