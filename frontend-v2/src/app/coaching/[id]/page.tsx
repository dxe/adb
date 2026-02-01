import { ContentWrapper } from '@/app/content-wrapper'
import { AuthedPageLayout } from '@/app/authed-page-layout'
import { EventForm } from '../../event/event-form'
import { Navbar } from '@/components/nav'

export default async function EditCoachingPage() {
  return (
    <AuthedPageLayout pageName="EditConnection_beta">
      <Navbar />
      <ContentWrapper size="sm" className="gap-8">
        <h1 className="text-3xl font-bold">Coaching</h1>
        <EventForm mode="connection" />
      </ContentWrapper>
    </AuthedPageLayout>
  )
}
