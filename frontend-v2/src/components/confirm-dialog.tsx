'use client'

import { ReactNode } from 'react'
import { AlertTriangle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'

export interface ConfirmDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  /** Main confirmation copy, rendered as a small paragraph. */
  children: ReactNode
  /** Optional copy rendered in an amber warning banner below the main copy. */
  warning?: ReactNode
  /** Label for the destructive confirm button. */
  confirmLabel: string
  /** Confirm button label while `isPending` (e.g. "Deleting..."). */
  pendingLabel: string
  cancelLabel?: string
  /** Called when the confirm button is clicked; callers own the mutation and cache invalidation. */
  onConfirm: () => void
  /** Disables both buttons and blocks dismissal while the action is in flight. */
  isPending: boolean
  /** Disables the confirm button, e.g. when a precondition for the action isn't met. */
  confirmDisabled?: boolean
}

/** Confirmation dialog shell for destructive actions (delete, hide, cancel). */
export function ConfirmDialog({
  open,
  onOpenChange,
  title,
  children,
  warning,
  confirmLabel,
  pendingLabel,
  cancelLabel = 'Cancel',
  onConfirm,
  isPending,
  confirmDisabled = false,
}: ConfirmDialogProps) {
  const handleOpenChange = (next: boolean) => {
    if (isPending) return
    onOpenChange(next)
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
        </DialogHeader>
        <p className="text-sm">{children}</p>
        {warning && (
          <div className="flex items-start gap-2 rounded-md border border-amber-300 bg-amber-50 p-3 text-sm text-amber-900">
            <AlertTriangle className="mt-0.5 h-4 w-4 shrink-0" />
            <p>{warning}</p>
          </div>
        )}
        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            onClick={() => handleOpenChange(false)}
            disabled={isPending}
          >
            {cancelLabel}
          </Button>
          <Button
            type="button"
            variant="destructive"
            onClick={onConfirm}
            disabled={isPending || confirmDisabled}
          >
            {isPending ? pendingLabel : confirmLabel}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
