import { ContentWrapper } from '@/app/content-wrapper'
import { EventForm } from '../../events/event-form'

export default function NewCoachingPage() {
  return (
    <ContentWrapper size="sm" className="gap-8">
      <h1 className="text-3xl font-bold">Coaching</h1>
      <EventForm mode="connection" />
    </ContentWrapper>
  )
}
