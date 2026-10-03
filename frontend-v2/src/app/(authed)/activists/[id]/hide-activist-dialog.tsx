'use client'

import { useMutation, useQueryClient } from '@tanstack/react-query'
import toast from 'react-hot-toast'
import { apiClient } from '@/lib/api'
import { activistKeys } from '@/lib/query-keys'
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
  activistId: number
  activistName: string
}

export function HideActivistDialog({
  open,
  onOpenChange,
  activistId,
  activistName,
}: Props) {
  const queryClient = useQueryClient()

  const mutation = useMutation({
    mutationFn: () => apiClient.hideActivist(activistId),
    onSuccess: () => {
      toast.success(`${activistName} was hidden`)
      // The activist drops out of the default (non-hidden) lists and counts,
      // and the detail page has to show the Hidden chip when it is revisited.
      queryClient.invalidateQueries({ queryKey: activistKeys.lists() })
      queryClient.invalidateQueries({ queryKey: activistKeys.counts() })
      queryClient.invalidateQueries({ queryKey: activistKeys.listBasic() })
      queryClient.invalidateQueries({
        queryKey: activistKeys.detail(activistId),
      })
      onOpenChange(false)
    },
    onError: (err: Error) => {
      toast.error(err.message || 'Failed to hide activist')
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
          <DialogTitle>Hide activist</DialogTitle>
        </DialogHeader>
        <p className="text-sm">
          WARNING: Hiding this activist will make them inaccessible unless they
          are unhidden by Tech. Are you sure you want to hide this activist?
        </p>
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
            disabled={mutation.isPending}
          >
            {mutation.isPending ? 'Hiding...' : 'Hide activist'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
