
import { AuthedPageLayout } from '@/app/authed-page-layout'
import { Navbar } from '@/components/nav'
import { ContentWrapper } from '@/app/content-wrapper'
import { UserForm } from '../user-form'

export default async function NewUserPage() {
  return (
    <AuthedPageLayout pageName="UserList">
      <Navbar />
      <ContentWrapper size="lg" className="gap-6">
        <UserForm />
      </ContentWrapper>
    </AuthedPageLayout>
  )
}
