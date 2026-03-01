import { Suspense } from 'react'
import { Loading } from '@/app/loading'
import { ContentWrapper } from '@/app/content-wrapper'
import { AuthedPageLayout } from '@/app/authed-page-layout'
import { Navbar } from '@/components/nav'
import { ActivistsSection } from './activists-section'

export default async function ActivistsPage() {
  return (
    <AuthedPageLayout pageName="TestPage">
      <Navbar />
      <ContentWrapper size="sm" className="gap-6">
        <p>Hello from App Router!</p>
        <p className="text-sm text-muted-foreground">
          This page streams the activists section so slow DB queries do not
          block the full page render.
        </p>
        <Suspense fallback={<Loading inline label="Loading activists..." />}>
          <ActivistsSection />
        </Suspense>
      </ContentWrapper>
    </AuthedPageLayout>
  )
}
