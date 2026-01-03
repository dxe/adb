import { ContentWrapper } from '@/app/content-wrapper'
import { AuthedPageLayout } from '@/app/authed-page-layout'
import { EventForm } from './event-form'
import { Navbar } from '@/components/nav'
import { Suspense } from 'react'

export default async function AttendancePage() {
  return (
    <AuthedPageLayout pageName="NewEvent_beta">
      <Navbar />
      <ContentWrapper size="md" className="gap-8">
        <div className="flex flex-col gap-3">
          <h1 className="text-3xl font-bold">Attendance</h1>
        </div>
        <Suspense fallback={<div>Loading form...</div>}>
          <EventForm />
        </Suspense>
      </ContentWrapper>
    </AuthedPageLayout>
  )
}
