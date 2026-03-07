import { ContentWrapper } from '@/app/content-wrapper'
import { Loading } from '@/app/loading'

export default function InterestGeneratorLoading() {
  return (
    <ContentWrapper size="sm" className="gap-6">
      <h1 className="text-lg">Interest Form Generator</h1>
      <Loading inline />
    </ContentWrapper>
  )
}
