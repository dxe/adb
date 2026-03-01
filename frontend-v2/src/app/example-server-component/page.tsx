import { Suspense } from 'react'
import { ContentWrapper } from '@/app/content-wrapper'
import { AuthedPageLayout } from '@/app/authed-page-layout'
import { Navbar } from '@/components/nav'
import { Loader2 } from 'lucide-react'
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
        <Suspense fallback={<ActivistsFallback />}>
          <ActivistsSection />
        </Suspense>
      </ContentWrapper>
    </AuthedPageLayout>
  )
}

const ActivistsFallback = () => (
  <div className="flex items-center gap-2 text-sm text-muted-foreground">
    <Loader2 className="h-4 w-4 animate-spin" aria-hidden="true" />
    <span>Loading activists...</span>
  </div>
)
