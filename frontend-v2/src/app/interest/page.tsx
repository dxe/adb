import { Suspense } from 'react'
import { ContentWrapper } from '@/app/content-wrapper'
import { InterestForm } from './interest-form'

// Public sign-up form ported from frontend/FormInterest.vue. Query params
// (built by the authed /interest/generator page) control which fields show.
export default function InterestPage() {
  return (
    <ContentWrapper size="sm" className="gap-6">
      <Suspense>
        <InterestForm />
      </Suspense>
    </ContentWrapper>
  )
}
