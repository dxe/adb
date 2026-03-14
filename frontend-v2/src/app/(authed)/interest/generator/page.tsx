import {
  dehydrate,
  HydrationBoundary,
  QueryClient,
} from '@tanstack/react-query'
import { ContentWrapper } from '@/app/content-wrapper'
import { API_PATH, ApiClient } from '@/lib/api'
import { getCookies } from '@/lib/auth'
import { redirectIfForbidden } from '@/lib/server-auth'
import GeneratorForm from './generator-form'

export default async function InterestGeneratorPage() {
  const apiClient = new ApiClient(await getCookies())
  const queryClient = new QueryClient()

  // Use fetchQuery instead of prefetchQuery so a 403 throws during SSR
  // and redirectIfForbidden can send the user to /403 immediately.
  await redirectIfForbidden(() =>
    queryClient.fetchQuery({
      queryKey: [API_PATH.CHAPTER_LIST],
      queryFn: ({ signal }) => apiClient.getChapterList(signal),
    }),
  )

  return (
    <ContentWrapper size="sm" className="gap-6">
      <HydrationBoundary state={dehydrate(queryClient)}>
        <h1 className="text-lg">Interest Form Generator</h1>
        <GeneratorForm />
      </HydrationBoundary>
    </ContentWrapper>
  )
}
