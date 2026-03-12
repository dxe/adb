import {
  dehydrate,
  HydrationBoundary,
  QueryClient,
} from '@tanstack/react-query'
import { ContentWrapper } from '@/app/content-wrapper'
import { ApiClient, CHAPTER_ORGANIZERS_QUERY_KEY } from '@/lib/api'
import { getCookies } from '@/lib/auth'
import OrganizersPage from './organizers-page'

export default async function IntlOrganizersPage() {
  const cookies = await getCookies()
  const queryClient = new QueryClient()
  const apiClient = new ApiClient(cookies)

  await queryClient.prefetchQuery({
    queryKey: [...CHAPTER_ORGANIZERS_QUERY_KEY],
    queryFn: ({ signal }) => apiClient.getChapterListWithOrganizers(signal),
  })

  return (
    <ContentWrapper size="full" className="gap-6">
      <HydrationBoundary state={dehydrate(queryClient)}>
        <OrganizersPage />
      </HydrationBoundary>
    </ContentWrapper>
  )
}
