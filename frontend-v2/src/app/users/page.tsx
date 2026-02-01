import { AuthedPageLayout } from '@/app/authed-page-layout'
import { Navbar } from '@/components/nav'
import { ContentWrapper } from '@/app/content-wrapper'
import UsersPage from './users-page'

export default async function UsersListPage() {
  return (
    <AuthedPageLayout pageName="UserList">
      <Navbar />
      <ContentWrapper size="xl" className="gap-6">
        <UsersPage />
      </ContentWrapper>
    </AuthedPageLayout>
  )
}
