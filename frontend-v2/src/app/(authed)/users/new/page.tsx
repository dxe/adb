import { ContentWrapper } from '@/app/content-wrapper'
import { UserForm } from '../user-form'

export default function NewUserPage() {
  return (
    <ContentWrapper size="lg" className="gap-6">
      <UserForm />
    </ContentWrapper>
  )
}
