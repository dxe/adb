import { ContentWrapper } from '@/app/content-wrapper'

export default function NotFoundPage() {
  return (
    <main className="flex min-h-[calc(100vh-8rem)] items-center justify-center">
      <ContentWrapper size="sm" className="items-center gap-4 text-center">
        <p className="text-sm font-medium uppercase tracking-[0.2em] text-neutral-500">
          404
        </p>
        <h1 className="text-3xl font-semibold text-neutral-900">
          Page not found
        </h1>
        <p className="text-balance text-neutral-600">
          The page you requested does not exist or is no longer available.
        </p>
        {/* Use a plain anchor so this points to the domain root instead of Next's /v2 base path. */}
        <a
          href="/"
          className="text-sm font-medium text-primary hover:underline"
        >
          Return home
        </a>
      </ContentWrapper>
    </main>
  )
}
