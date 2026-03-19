import { ContentWrapper } from '@/app/content-wrapper'
import { Loading } from '@/app/loading'

export default function ActivistLoading() {
  return (
    <ContentWrapper size="lg" className="gap-6">
      <Loading inline />
    </ContentWrapper>
  )
}
