import { ContentWrapper } from '@/app/content-wrapper'
import { Loading } from '@/app/loading'

export default function UsersLoading() {
  return (
    <ContentWrapper size="xl" className="gap-6">
      <h1 className="text-2xl font-semibold">Users</h1>
      <Loading inline />
    </ContentWrapper>
  )
}
