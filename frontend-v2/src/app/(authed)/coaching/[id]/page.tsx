import { notFound } from 'next/navigation'
import { ContentWrapper } from '@/app/content-wrapper'
import { EventForm } from '../../event/event-form'

export default async function EditCoachingPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = await params
  const eventId = parseInt(id)
  if (Number.isNaN(eventId)) {
    notFound()
  }

  return (
    <ContentWrapper size="sm" className="gap-8">
      <h1 className="text-3xl font-bold">Coaching</h1>
      <EventForm mode="connection" />
    </ContentWrapper>
  )
}
