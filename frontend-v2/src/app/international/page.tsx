import { Metadata } from 'next'
import { ContentWrapper } from '@/app/content-wrapper'
import { InternationalForm } from './international-form'

export const metadata: Metadata = {
  title: 'Join DxE',
}

export default function InternationalPage() {
  return (
    <ContentWrapper size="md" className="gap-6">
      <h1 className="text-lg">Sign up to join our International Network</h1>
      <InternationalForm />
    </ContentWrapper>
  )
}
