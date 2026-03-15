import { ContentWrapper } from '@/app/content-wrapper'

export default function ForbiddenPage() {
  return (
    <main className="flex min-h-[calc(100vh-8rem)] items-center justify-center">
      <ContentWrapper size="sm" className="items-center gap-4 text-center">
        <p className="text-sm font-medium uppercase tracking-[0.2em] text-neutral-500">
          403
        </p>
        <h1 className="text-3xl font-semibold text-neutral-900">
          Access denied
        </h1>
        <p className="text-balance text-neutral-600">
          You do not have permission to view this page.
        </p>
      </ContentWrapper>
    </main>
  )
}
