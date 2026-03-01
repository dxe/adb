import { ContentWrapper } from '@/app/content-wrapper'
import UsersPage from './users-page'

export default function UsersListPage() {
  return (
    <ContentWrapper size="xl" className="gap-6">
      <UsersPage />
    </ContentWrapper>
  )
}
