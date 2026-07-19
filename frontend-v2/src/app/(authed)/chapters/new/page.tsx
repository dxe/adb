import { forbidden } from 'next/navigation'
import { ContentWrapper } from '@/app/content-wrapper'
import { getCachedSession } from '@/app/session'
import { ChapterForm } from '../chapter-form'

export default async function NewChapterPage() {
  const session = await getCachedSession()
  // Mirrors the Go apiIntlCoordinatorAuthMiddleware role check.
  const roles = session.user?.Roles ?? []
  if (!roles.includes('admin') && !roles.includes('intl_coordinator')) {
    forbidden()
  }

  return (
    <ContentWrapper size="lg" className="gap-6">
      <ChapterForm />
    </ContentWrapper>
  )
}
