import {
  dehydrate,
  HydrationBoundary,
  QueryClient,
} from '@tanstack/react-query'
import { ContentWrapper } from '@/app/content-wrapper'
import { AuthedPageLayout } from '@/app/authed-page-layout'
import { Navbar } from '@/components/nav'
import GeneratorForm from './generator-form'
import { API_PATH, ApiClient } from '@/lib/api'
import { getCookies } from '@/lib/auth'

export default async function InterestGeneratorPage() {
  const apiClient = new ApiClient(await getCookies())
  const queryClient = new QueryClient()

  // Prefetch chapter list for form
  await queryClient.prefetchQuery({
    queryKey: [API_PATH.CHAPTER_LIST],
    queryFn: apiClient.getChapterList,
  })

  return (
    <AuthedPageLayout pageName="InterestFormGenerator">
      <Navbar />
      <ContentWrapper size="sm" className="gap-6">
        <HydrationBoundary state={dehydrate(queryClient)}>
          <h1 className="text-lg">Interest Form Generator</h1>
          <GeneratorForm />
        </HydrationBoundary>
      </ContentWrapper>
    </AuthedPageLayout>
  )
}
