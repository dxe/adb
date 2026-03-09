import { ContentWrapper } from '@/app/content-wrapper'
import { EventsPageLoader } from './events-page-loader'

export default function EventsListPage() {
  return (
    <ContentWrapper size="xl" className="gap-6">
      <EventsPageLoader />
    </ContentWrapper>
  )
}
