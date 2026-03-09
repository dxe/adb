import { ContentWrapper } from '@/app/content-wrapper'
import { Loading } from '@/app/loading'

export default function EditCoachingLoading() {
  return (
    <ContentWrapper size="sm" className="gap-8">
      <h1 className="text-3xl font-bold">Coaching</h1>
      <Loading inline />
    </ContentWrapper>
  )
}
