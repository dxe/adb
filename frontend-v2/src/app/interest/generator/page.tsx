// Modeled from example from React Query docs:
// https://tanstack.com/query/latest/docs/framework/react/guides/advanced-ssr

import {
  dehydrate,
  HydrationBoundary,
  QueryClient,
} from '@tanstack/react-query'
import { ContentWrapper } from '@/app/content-wrapper'
import { AuthedPageLayout } from '@/app/authed-page-layout'
import { API_PATH, ApiClient } from '@/lib/api'
import { getCookies } from '@/lib/auth'
import { Navbar } from '@/components/nav'

import GeneratorForm from './generator-form'

export default async function InterestGeneratorPage() {
  const apiClient = new ApiClient(await getCookies())
  const queryClient = new QueryClient()

  await queryClient.prefetchQuery({
    queryKey: [API_PATH.CHAPTER_LIST],
    queryFn: apiClient.getChapterList,
  })

  return (
    <AuthedPageLayout pageName="InterestFormGenerator">
      <HydrationBoundary state={dehydrate(queryClient)}>
        <Navbar />
        <ContentWrapper size="sm" className="gap-6">
          <h1 className="text-lg">Interest Form Generator</h1>

          <GeneratorForm />
        </ContentWrapper>
      </HydrationBoundary>
    </AuthedPageLayout>
  )
}
