'use client'

import { useMutation, useQueryClient } from '@tanstack/react-query'
import toast from 'react-hot-toast'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { API_PATH, apiClient, WorkingGroup } from '@/lib/api'

export function DeleteWorkingGroupDialog({
  workingGroup,
  onClose,
}: {
  workingGroup: WorkingGroup
  onClose: () => void
}) {
  const queryClient = useQueryClient()

  // Best-effort mirror of the backend rule: working_group/delete rejects
  // groups that still have working_group_members rows.
  const hasMembers = workingGroup.members.length > 0

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
    <ConfirmDialog
      open
      onOpenChange={(open) => !open && onClose()}
      title="Delete working group"
      warning={
        hasMembers
          ? 'This working group still has members (including non-members on its mailing list). Remove them all before it can be deleted.'
          : 'Before deleting a working group, be sure to remove all members of that group.'
      }
      confirmLabel="Delete"
      pendingLabel="Deleting..."
      onConfirm={() => mutation.mutate()}
      isPending={mutation.isPending}
      confirmDisabled={hasMembers}
    >
      Are you sure you want to delete <strong>{workingGroup.name}</strong>?
    </ConfirmDialog>
  )
}
