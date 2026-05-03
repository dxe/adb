'use client'

import { useRouter } from 'next/navigation'
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
  activistId: number
  activistName: string
}

export function HideActivistDialog({
  open,
  onOpenChange,
  activistId,
  activistName,
}: Props) {
  const router = useRouter()
  const queryClient = useQueryClient()

  const mutation = useMutation({
    mutationFn: () => apiClient.hideActivist(activistId),
    onSuccess: () => {
      toast.success(`${activistName} was hidden`)
      queryClient.invalidateQueries({ queryKey: [API_PATH.ACTIVISTS_SEARCH] })
      queryClient.invalidateQueries({
        queryKey: [API_PATH.ACTIVIST_LIST_BASIC],
      })
      onOpenChange(false)
      router.push('/activists')
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
