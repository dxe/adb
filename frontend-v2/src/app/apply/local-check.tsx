'use client'

import { Button } from '@/components/ui/button'

/**
 * Gate shown before anything else: chapter membership is specific to the
 * SF Bay Area chapter, so we ask if the applicant lives nearby before
 * showing the rest of the form. Matches the `local` radio buttons in
 * frontend/FormApply.vue.
 */
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
