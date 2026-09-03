'use client'

import { Button } from '@/components/ui/button'

/** Asks whether the applicant lives near the SF Bay Area chapter before showing the rest of the flow. */
export function LocalCheck({
  onLocal,
  onNotLocal,
}: {
  onLocal: () => void
  onNotLocal: () => void
}) {
  return (
    <div className="flex flex-col gap-2">
      <p className="font-medium">
        Do you live within 100 miles of Berkeley, CA?
      </p>
      <div className="flex gap-3">
        <Button type="button" onClick={onLocal}>
          Yes
        </Button>
        <Button type="button" variant="outline" onClick={onNotLocal}>
          No
        </Button>
      </div>
    </div>
  )
}
