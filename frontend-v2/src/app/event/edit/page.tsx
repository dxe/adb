import { ContentWrapper } from '@/app/content-wrapper'
import { AuthedPageLayout } from '@/app/authed-page-layout'
import { EventForm } from './event-form'
import { Navbar } from '@/components/nav'

export default async function AttendancePage() {
  return (
    <AuthedPageLayout pageName="NewEvent_beta">
      <Navbar />
      <ContentWrapper size="md" className="gap-8">
        <div className="flex flex-col gap-3">
          <h1 className="text-3xl font-bold">Attendance</h1>
          <p className="text-neutral-500">
            Enter the names of attendees below. Type to see suggestions or enter
            new names.
          </p>
          <p className="font-semibold text-red-500">
            Note that this page is a work in progress and may not actually
            create an event!
          </p>
        </div>
        <EventForm />
      </ContentWrapper>
    </AuthedPageLayout>
  )
}
