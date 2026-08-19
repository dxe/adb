import type { Metadata } from 'next'
import { ContentWrapper } from '@/app/content-wrapper'
import { ApplyForm } from '@/app/apply/apply-form'

export const metadata: Metadata = {
  title: 'Join DxE SF Bay',
}

export default function ApplyPage() {
  return (
    <ContentWrapper size="md" className="gap-6">
      <h1 className="text-2xl font-bold">Join the SF Bay Area Chapter</h1>
      <ApplyForm />
    </ContentWrapper>
  )
}
