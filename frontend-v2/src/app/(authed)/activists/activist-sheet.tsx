// Defines the ActivistSheet component which allows editing an activist without
// leaving whatever page the user is already on. "Sheet" here refers to a
// UX/UI concept (shadcn/ui Sheet, built on Radix Dialog), not to be confused
// with any spreadsheet-like table.

'use client'

import * as DialogPrimitive from '@radix-ui/react-dialog'
import Link from 'next/link'
import { ExternalLink, X } from 'lucide-react'
import { useState } from 'react'
import { ActivistDetail } from './[id]/activist-detail'

interface ActivistSheetProps {
  activistId: number | null
  onClose: () => void
}

export function ActivistSheet({ activistId, onClose }: ActivistSheetProps) {
  // Holds the last non-null activistId. Never cleared back to null so content
  // stays mounted and visible while Radix plays the exit animation.
  const [displayActivistId, setDisplayActivistId] = useState<number | null>(
    null,
  )
  const [prevActivistId, setPrevActivistId] = useState<number | null>(null)

  // React derived-state pattern: update synchronously during render so content
  // is present on the first open paint without waiting for an effect.
  if (activistId !== null && activistId !== prevActivistId) {
    setPrevActivistId(activistId)
    setDisplayActivistId(activistId)
  }

  return (
    <DialogPrimitive.Root
      modal={false}
      open={activistId !== null}
      onOpenChange={(open) => {
        if (!open) onClose()
      }}
    >
      <DialogPrimitive.Portal>
        <DialogPrimitive.Content className="fixed inset-y-0 right-0 z-50 flex h-full w-full flex-col overflow-hidden border-l bg-background shadow-lg transition ease-in-out data-[state=closed]:duration-300 data-[state=open]:duration-500 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:slide-out-to-right data-[state=open]:slide-in-from-right sm:max-w-2xl">
          <DialogPrimitive.Title className="sr-only">
            Activist Details
          </DialogPrimitive.Title>
          <DialogPrimitive.Description className="sr-only">
            View and edit activist information
          </DialogPrimitive.Description>

          <div className="flex items-center justify-between border-b px-6 py-3">
            {displayActivistId !== null && (
              <Link
                href={`/activists/${displayActivistId}`}
                className="flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground"
              >
                <ExternalLink className="h-4 w-4" />
                Full page
              </Link>
            )}
            <DialogPrimitive.Close className="ml-auto rounded-sm opacity-70 ring-offset-background transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2">
              <X className="h-4 w-4" />
              <span className="sr-only">Close</span>
            </DialogPrimitive.Close>
          </div>

          <div className="flex flex-1 flex-col gap-6 overflow-y-auto px-6 pb-6 pt-4">
            {displayActivistId !== null && (
              <ActivistDetail activistId={displayActivistId} />
            )}
          </div>
        </DialogPrimitive.Content>
      </DialogPrimitive.Portal>
    </DialogPrimitive.Root>
  )
}
