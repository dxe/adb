import { ContentWrapper } from '@/app/content-wrapper'
import { Loading } from '@/app/loading'

export default function NewUserLoading() {
  return (
    <ContentWrapper size="lg" className="gap-6">
      <div className="flex flex-col gap-1">
        <p className="text-sm uppercase tracking-wide text-muted-foreground">
          Create user
        </p>
        <Loading inline />
      </div>
    </ContentWrapper>
  )
}
