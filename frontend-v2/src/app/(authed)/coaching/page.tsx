import { ContentWrapper } from '@/app/content-wrapper'
import { EventForm } from '../event/event-form'

export default function CoachingPage() {
  return (
    <ContentWrapper size="sm" className="gap-8">
      <h1 className="text-3xl font-bold">Coaching</h1>
      <EventForm mode="connection" />
    </ContentWrapper>
  )
}
