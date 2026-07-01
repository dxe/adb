import { ContentWrapper } from '@/app/content-wrapper'
import { HomeHub } from './home-hub'

export default function HomePage() {
  return (
    <ContentWrapper size="md" className="gap-6">
      <HomeHub />
    </ContentWrapper>
  )
}
