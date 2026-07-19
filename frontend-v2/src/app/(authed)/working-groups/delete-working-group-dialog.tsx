'use client'

import { useMutation, useQueryClient } from '@tanstack/react-query'
import toast from 'react-hot-toast'
import { AlertTriangle, Loader2 } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { API_PATH, apiClient, WorkingGroup } from '@/lib/api'

export function DeleteWorkingGroupDialog({
  workingGroup,
  onClose,
}: {
  workingGroup: WorkingGroup
  onClose: () => void
}) {
  const queryClient = useQueryClient()

  const mutation = useMutation({
    mutationFn: () => apiClient.deleteWorkingGroup(workingGroup.id),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: [API_PATH.WORKING_GROUP_LIST],
      })
      toast.success(`${workingGroup.name} deleted`)
      onClose()
    },
    onError: (err: unknown) => {
      const message =
        err instanceof Error ? err.message : 'Failed to delete working group'
      toast.error(`Error: ${message}`)
    },
  })

  return (
    <Dialog open onOpenChange={(open) => !open && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Delete working group</DialogTitle>
        </DialogHeader>
        <p className="text-sm">
          Are you sure you want to delete <strong>{workingGroup.name}</strong>?
        </p>
        <div className="flex items-start gap-2 rounded-md border border-yellow-300 bg-yellow-50 p-3 text-sm text-yellow-900">
          <AlertTriangle className="mt-0.5 h-4 w-4 shrink-0" />
          <span>
            Before deleting a working group, be sure to remove all members of
            that group.
          </span>
        </div>
        <DialogFooter>
          <Button
            variant="outline"
            onClick={onClose}
            disabled={mutation.isPending}
          >
            Cancel
          </Button>
          <Button
            variant="destructive"
            onClick={() => mutation.mutate()}
            disabled={mutation.isPending}
          >
            {mutation.isPending && <Loader2 className="h-4 w-4 animate-spin" />}
            Delete
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
