import { ContentWrapper } from '@/app/content-wrapper'
import { Loading } from '@/app/loading'

export default function ActivistsLoading() {
  return (
    <ContentWrapper size="xl" className="gap-6">
      <h1 className="text-2xl font-semibold">Activists</h1>
      <Loading inline />
    </ContentWrapper>
  )
}
