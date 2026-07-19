'use client'

import { useMutation, useQueryClient } from '@tanstack/react-query'
import toast from 'react-hot-toast'
import { AlertTriangle } from 'lucide-react'
import { API_PATH, apiClient, CircleGroup } from '@/lib/api'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  circle: CircleGroup
}

export function DeleteCircleDialog({ open, onOpenChange, circle }: Props) {
  const queryClient = useQueryClient()

  // Best-effort mirror of the backend rule: circle/delete rejects circles that still have members.
  const hasMembers = circle.members.length > 0

  const mutation = useMutation({
    mutationFn: () => apiClient.deleteCircle(circle.id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [API_PATH.CIRCLE_LIST] })
      toast.success(`${circle.name} deleted`)
      onOpenChange(false)
    },
    onError: (err: Error) => {
      toast.error(err.message || 'Failed to delete circle. Please try again.')
    },
  })

  const handleOpenChange = (next: boolean) => {
    if (mutation.isPending) return
    onOpenChange(next)
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Delete circle</DialogTitle>
        </DialogHeader>
        <p className="text-sm">
          Are you sure you want to delete <strong>{circle.name}</strong>?
        </p>
        <div className="flex items-start gap-2 rounded-md border border-amber-300 bg-amber-50 p-3 text-sm text-amber-900">
          <AlertTriangle className="mt-0.5 h-4 w-4 shrink-0" />
          <p>
            {hasMembers
              ? 'This circle still has members. Remove them all before it can be deleted.'
              : 'Before deleting a circle, be sure to remove all members of that circle.'}
          </p>
        </div>
        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            onClick={() => handleOpenChange(false)}
            disabled={mutation.isPending}
          >
            Cancel
          </Button>
          <Button
            type="button"
            variant="destructive"
            onClick={() => mutation.mutate()}
            disabled={mutation.isPending || hasMembers}
          >
            {mutation.isPending ? 'Deleting...' : 'Delete'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
