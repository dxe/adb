import { AuthedPageLayout } from '@/app/authed-page-layout'
import { Navbar } from '@/components/nav'
import { ContentWrapper } from '@/app/content-wrapper'
import { UserForm } from '../user-form'

export default async function EditUserPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = await params
  const userId = Number(id)

  return (
    <AuthedPageLayout pageName="UserList">
      <Navbar />
      <ContentWrapper size="lg" className="gap-6">
        <UserForm userId={userId} />
      </ContentWrapper>
    </AuthedPageLayout>
  )
}
