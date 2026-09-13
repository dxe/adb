import { ContentWrapper } from '@/app/content-wrapper'
import { Loading } from '@/app/loading'

export default function CirclesLoading() {
  return (
    <ContentWrapper size="xl" className="gap-6">
      <h1 className="text-2xl font-semibold">Circles</h1>
      <Loading inline />
    </ContentWrapper>
  )
}
