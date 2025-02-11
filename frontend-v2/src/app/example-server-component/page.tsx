// Modeled from example from React Query docs:
// https://tanstack.com/query/latest/docs/framework/react/guides/advanced-ssr

import {
  dehydrate,
  HydrationBoundary,
  QueryClient,
} from '@tanstack/react-query'
import Activists, { getActivists } from './activists'
import { VueNavbar } from 'app/VueNavbar'
import { ContentWrapper } from 'app/ContentWrapper'
import { AuthedPageLayout } from 'app/AuthedPageLayout'

export default async function ActivistsPage() {
  const queryClient = new QueryClient()

  await queryClient.prefetchQuery({
    queryKey: ['activists'],
    queryFn: getActivists,
  })

  return (
    <AuthedPageLayout>
      <VueNavbar pageName="TestPage" />
      <ContentWrapper size="sm" className="gap-6">
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
