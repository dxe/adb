import { Loader2 } from 'lucide-react'
import { cn } from '@/lib/utils'

type LoadingProps = {
  label?: string
  inline?: boolean
  className?: string
}

export function Loading({
  label = 'Loading page...',
  inline = false,
  className,
}: LoadingProps) {
  return (
    <div
      className={cn(
        inline
          ? 'flex items-center gap-2 text-sm text-muted-foreground'
          : 'flex min-h-[calc(100vh-6.25rem)] items-center justify-center px-4',
        className,
      )}
    >
      <div
        className={cn(
          'flex items-center',
          inline
            ? 'gap-2 text-sm text-muted-foreground'
            : 'gap-3 rounded-md bg-white/90 px-4 py-3 text-sm text-neutral-700 shadow-lg backdrop-blur-sm',
        )}
        role="status"
        aria-live="polite"
      >
        <Loader2
          className="h-5 w-5 animate-spin text-primary"
          aria-hidden="true"
        />
        <span>{label}</span>
      </div>
    </div>
  )
}

export default function RouteLoading() {
  return <Loading />
}
