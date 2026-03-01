import { ContentWrapper } from '@/app/content-wrapper'
import { EventForm } from './event-form'

export default async function AttendancePage() {
  return (
    <ContentWrapper size="sm" className="gap-8">
      <h1 className="text-3xl font-bold">Attendance</h1>
      <EventForm mode="event" />
    </ContentWrapper>
  )
}
