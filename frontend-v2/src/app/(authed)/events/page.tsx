import { ContentWrapper } from '@/app/content-wrapper'
import EventsPage from './events-page'

export default function EventsListPage() {
  return (
    <ContentWrapper size="xl" className="gap-6">
      <EventsPage />
    </ContentWrapper>
  )
}
