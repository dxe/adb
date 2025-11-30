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
import { Navbar } from '@/components/nav'

export default async function ActivistsPage() {
  const apiClient = new ApiClient(await getCookies())
  const queryClient = new QueryClient()

  await queryClient.prefetchQuery({
    queryKey: [API_PATH.ACTIVIST_NAMES_GET],
    queryFn: apiClient.getActivistNames,
  })

  // For navbar. TODO: only do if user is admin.
  // https://app.asana.com/1/71341131816665/project/1209217418568645/task/1212207774751674
  await queryClient.prefetchQuery({
    queryKey: [API_PATH.CHAPTER_LIST],
    queryFn: apiClient.getChapterList,
  })

  return (
    <AuthedPageLayout pageName="TestPage">
      {
        // Serialization is as easy as passing props.
        // HydrationBoundary is a Client Component, so hydration will happen there.
      }
      <HydrationBoundary state={dehydrate(queryClient)}>
        <Navbar />
        <ContentWrapper size="sm" className="gap-6">
          <p>Hello from App Router!</p>

          <Activists />
        </ContentWrapper>
      </HydrationBoundary>
    </AuthedPageLayout>
  )
}
