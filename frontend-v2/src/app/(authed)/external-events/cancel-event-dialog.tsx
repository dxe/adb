'use client'

import { useMutation, useQueryClient } from '@tanstack/react-query'
import toast from 'react-hot-toast'
import { API_PATH, apiClient } from '@/lib/api'
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
      toast.success('Successfully cancelled event.')
      queryClient.invalidateQueries({
        queryKey: [API_PATH.EXTERNAL_EVENTS_LIST],
      })
      onOpenChange(false)
    },
    onError: (err: Error) => {
      toast.error(err.message || 'Failed to cancel event')
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
          <DialogTitle>Cancel event</DialogTitle>
        </DialogHeader>
        <p className="text-sm">
          Are you sure you want to cancel &ldquo;{eventName}&rdquo;? It will no
          longer be displayed on the public events page.
        </p>
        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            onClick={() => handleOpenChange(false)}
            disabled={mutation.isPending}
          >
            Keep event
          </Button>
          <Button
            type="button"
            variant="destructive"
            onClick={() => mutation.mutate()}
            disabled={mutation.isPending}
          >
            {mutation.isPending ? 'Cancelling...' : 'Cancel event'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
