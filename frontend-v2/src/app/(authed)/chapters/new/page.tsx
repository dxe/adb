import { ContentWrapper } from '@/app/content-wrapper'
import { ApiClient } from '@/lib/api'
import { getCookies } from '@/lib/auth'
import { redirectForHttpError } from '@/lib/server-auth'
import { ChapterForm } from '../chapter-form'

export default async function NewChapterPage() {
  const apiClient = new ApiClient(await getCookies())

  // Not used by the form itself; this call exists purely to gate the page
  // behind the same admin/intl_coordinator check as the rest of /chapters.
  await redirectForHttpError(() => apiClient.getChapterAdminList())

  return (
    <ContentWrapper size="lg" className="gap-6">
      <ChapterForm />
    </ContentWrapper>
  )
}
