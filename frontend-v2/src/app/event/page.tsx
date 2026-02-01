
import { ContentWrapper } from '@/app/content-wrapper'
import { AuthedPageLayout } from '@/app/authed-page-layout'
import { EventForm } from './event-form'
import { Navbar } from '@/components/nav'

export default async function AttendancePage() {
  return (
    <AuthedPageLayout pageName="NewEvent_beta">
      <Navbar />
      <ContentWrapper size="sm" className="gap-8">
        <h1 className="text-3xl font-bold">Attendance</h1>
        <EventForm mode="event" />
      </ContentWrapper>
    </AuthedPageLayout>
  )
}
