import { ContentWrapper } from '@/app/content-wrapper'
import { AuthedPageLayout } from '@/app/authed-page-layout'
import { EventForm } from '../event/event-form'
import { Navbar } from '@/components/nav'
import { Suspense } from 'react'

export default async function CoachingPage() {
  return (
    <AuthedPageLayout pageName="NewConnection_beta">
      <Navbar />
      <ContentWrapper size="md" className="gap-8">
        <h1 className="text-3xl font-bold">Coaching</h1>
        <Suspense fallback={<div>Loading form...</div>}>
          <EventForm mode="connection" />
        </Suspense>
      </ContentWrapper>
    </AuthedPageLayout>
  )
}
