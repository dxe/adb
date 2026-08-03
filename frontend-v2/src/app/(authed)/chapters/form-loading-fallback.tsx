import { Loader2 } from 'lucide-react'

export function FormLoadingFallback() {
  return (
    <div className="flex items-center gap-2 text-muted-foreground text-sm">
      <Loader2 className="h-4 w-4 animate-spin" />
      Loading...
    </div>
  )
}
