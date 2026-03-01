import { ContentWrapper } from '@/app/content-wrapper'
import { Loading } from '@/app/loading'

export default function AuthedLoading() {
  return (
    <ContentWrapper size="xl" className="gap-6">
      <Loading inline />
    </ContentWrapper>
  )
}
