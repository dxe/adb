import { forbidden } from 'next/navigation'
import { ContentWrapper } from '@/app/content-wrapper'
import { getCachedSession } from '@/app/session'
import ConfigPage from './config-page'

export default async function ConfigurationPage() {
  const session = await getCachedSession()
  if (!session.user?.Roles.includes('admin')) {
    forbidden()
  }

  return (
    <ContentWrapper size="xl" className="gap-6">
      <ConfigPage />
    </ContentWrapper>
  )
}
