'use client'

import { useMutation, useQueryClient } from '@tanstack/react-query'
import toast from 'react-hot-toast'
import { API_PATH, apiClient, EventListItem } from '@/lib/api'
import { ConfirmDialog } from '@/components/confirm-dialog'

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  event: EventListItem
}

export function DeleteEventDialog({ open, onOpenChange, event }: Props) {
  const queryClient = useQueryClient()
  const noun = event.event_type === 'Connection' ? 'coaching' : 'event'

  const mutation = useMutation({
    mutationFn: () => apiClient.deleteEvent(event.event_id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [API_PATH.EVENT_LIST] })
      toast.success(`${event.event_name} deleted`)
      onOpenChange(false)
    },
    onError: (err: Error) => {
      toast.error(err.message || `Failed to delete ${noun}. Please try again.`)
    },
  })

  return (
    <ConfirmDialog
      open={open}
      onOpenChange={onOpenChange}
      title={`Delete ${noun}`}
      confirmLabel="Delete"
      pendingLabel="Deleting..."
      onConfirm={() => mutation.mutate()}
      isPending={mutation.isPending}
    >
      Are you sure you want to delete <strong>{event.event_name}</strong>?
    </ConfirmDialog>
  )
}
