import { notFound } from 'next/navigation'
import { ContentWrapper } from '@/app/content-wrapper'
import { UserForm } from '../user-form'

export default async function EditUserPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = await params
  const userId = parseInt(id)
  if (Number.isNaN(userId)) {
    notFound()
  }

  return (
    <ContentWrapper size="lg" className="gap-6">
      <UserForm userId={userId} />
    </ContentWrapper>
  )
}
