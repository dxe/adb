import { ContentWrapper } from '@/app/content-wrapper'
import { Loading } from '@/app/loading'

export default function EditUserLoading() {
  return (
    <ContentWrapper size="lg" className="gap-6">
      <div className="flex flex-col gap-1">
        <p className="text-sm uppercase tracking-wide text-muted-foreground">
          Edit user
        </p>
        <Loading inline />
      </div>
    </ContentWrapper>
  )
}
