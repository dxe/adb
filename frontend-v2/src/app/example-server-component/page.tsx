// Modeled from example from React Query docs:
// https://tanstack.com/query/latest/docs/framework/react/guides/advanced-ssr

import {
  dehydrate,
  HydrationBoundary,
  QueryClient,
} from '@tanstack/react-query'
import Activists, { getActivists } from './activists'

export default async function ActivistsPage() {
  const queryClient = new QueryClient()

  await queryClient.prefetchQuery({
    queryKey: ['activists'],
    queryFn: getActivists,
  })

  return (
    // Serialization is as easy as passing props.
    // HydrationBoundary is a Client Component, so hydration will happen there.
    <HydrationBoundary state={dehydrate(queryClient)}>
      <Activists />
    </HydrationBoundary>
  )
}
