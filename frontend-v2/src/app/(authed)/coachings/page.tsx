import { ContentWrapper } from '@/app/content-wrapper'
import { CoachingListPageLoader } from './coaching-list-page-loader'

export default function CoachingPage() {
  return (
    <ContentWrapper size="xl" className="gap-6">
      <CoachingListPageLoader />
    </ContentWrapper>
  )
}
