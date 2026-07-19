import type { Metadata } from 'next'
import { ContentWrapper } from '@/app/content-wrapper'
import { ApplyForm } from '@/app/apply/apply-form'

export const metadata: Metadata = {
  title: 'Join DxE SF Bay',
}

// Public, unauthenticated page (no navbar / auth provider). Mirrors the
// legacy Vue page at frontend/FormApply.vue, served by the Go backend at
// the (unauthenticated) `/apply` route.
export default function ApplyPage() {
  return (
    <ContentWrapper size="md" className="gap-6">
      <h1 className="text-2xl font-bold">Join the SF Bay Area Chapter</h1>
      <ApplyForm />
    </ContentWrapper>
  )
}
