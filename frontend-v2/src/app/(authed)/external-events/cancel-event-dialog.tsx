'use client'

import { useMutation, useQueryClient } from '@tanstack/react-query'
import toast from 'react-hot-toast'
import { API_PATH, apiClient } from '@/lib/api'
import { ConfirmDialog } from '@/components/confirm-dialog'

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  eventId: string
  eventName: string
}

export function CancelEventDialog({
  open,
  onOpenChange,
  eventId,
  eventName,
}: Props) {
  const queryClient = useQueryClient()

  const mutation = useMutation({
    mutationFn: () => apiClient.cancelExternalEvent(eventId),
    onSuccess: () => {
      toast.success('Successfully deleted event.')
      queryClient.invalidateQueries({
        queryKey: [API_PATH.EXTERNAL_EVENTS_LIST],
      })
      onOpenChange(false)
    },
    onError: (err: Error) => {
      toast.error(err.message || 'Failed to delete event')
    },
  })

  return (
    <ConfirmDialog
      open={open}
      onOpenChange={onOpenChange}
      title="Delete event"
      confirmLabel="Delete event"
      pendingLabel="Deleting..."
      cancelLabel="Keep event"
      onConfirm={() => mutation.mutate()}
      isPending={mutation.isPending}
    >
      Are you sure you want to delete &ldquo;{eventName}&rdquo;? It will no
      longer be displayed on the public events page.
    </ConfirmDialog>
  )
}
