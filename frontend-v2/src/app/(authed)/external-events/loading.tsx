import { ContentWrapper } from '@/app/content-wrapper'
import { Loading } from '@/app/loading'

export default function ExternalEventsLoading() {
  return (
    <ContentWrapper size="xl" className="gap-6">
      <h1 className="text-2xl font-semibold">Facebook Events</h1>
      <Loading inline />
    </ContentWrapper>
  )
}
