import { ContentWrapper } from '@/app/content-wrapper'
import CoachingListPage from './coaching-list-page'

export default function CoachingPage() {
  return (
    <ContentWrapper size="xl" className="gap-6">
      <CoachingListPage />
    </ContentWrapper>
  )
}
