import { ContentWrapper } from '@/app/content-wrapper'
import { Loading } from '@/app/loading'

export default function ChaptersLoading() {
  return (
    <ContentWrapper size="full" className="gap-6">
      <h1 className="text-2xl font-semibold">Chapters</h1>
      <Loading inline />
    </ContentWrapper>
  )
}
