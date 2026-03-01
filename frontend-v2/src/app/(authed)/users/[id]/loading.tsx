import { ContentWrapper } from '@/app/content-wrapper'
import { Loading } from '@/app/loading'

export default function EditUserLoading() {
  return (
    <ContentWrapper size="lg" className="gap-6">
      <h1 className="text-2xl font-semibold">User</h1>
      <Loading inline />
    </ContentWrapper>
  )
}
