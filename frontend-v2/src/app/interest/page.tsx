import { Suspense } from 'react'
import { ContentWrapper } from '@/app/content-wrapper'
import { InterestForm } from './interest-form'

// Public interest / check-in forms. Query params (built by the
// /interest/generator page) control which fields show.
export default function InterestPage() {
  return (
    <ContentWrapper size="sm" className="gap-6">
      <Suspense>
        <InterestForm />
      </Suspense>
    </ContentWrapper>
  )
}
