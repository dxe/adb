'use client'

import { useMutation, useQueryClient } from '@tanstack/react-query'
import toast from 'react-hot-toast'
import { API_PATH, apiClient, CircleGroup } from '@/lib/api'
import { ConfirmDialog } from '@/components/confirm-dialog'

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

  return (
    <ConfirmDialog
      open={open}
      onOpenChange={onOpenChange}
      title="Delete circle"
      warning={
        hasMembers
          ? 'This circle still has members. Remove them all before it can be deleted.'
          : 'Before deleting a circle, be sure to remove all members of that circle.'
      }
      confirmLabel="Delete"
      pendingLabel="Deleting..."
      onConfirm={() => mutation.mutate()}
      isPending={mutation.isPending}
      confirmDisabled={hasMembers}
    >
      Are you sure you want to delete <strong>{circle.name}</strong>?
    </ConfirmDialog>
  )
}
