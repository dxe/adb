import { Loader2 } from 'lucide-react'

export default function Loading() {
  return (
    <div className="flex min-h-[calc(100vh-6.25rem)] items-center justify-center px-4">
      <div className="flex items-center gap-3 rounded-md bg-white/90 px-4 py-3 text-sm text-neutral-700 shadow-lg backdrop-blur-sm">
        <Loader2
          className="h-5 w-5 animate-spin text-primary"
          aria-hidden="true"
        />
        <span>Loading page...</span>
      </div>
    </div>
  )
}
