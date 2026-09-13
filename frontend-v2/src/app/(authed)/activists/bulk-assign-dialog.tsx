'use client'

import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import toast from 'react-hot-toast'
import { API_PATH, apiClient } from '@/lib/api'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Label } from '@/components/ui/label'

// The server treats an assignee of 0 as "unassign".
const UNASSIGNED_USER_ID = 0

// Mirrors model.MaxBulkAssignActivists: the server rejects a larger set rather
// than truncating it, so catch it here with a message naming the limit.
const MAX_BULK_ASSIGN = 1000

interface BulkAssignDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  activistIds: number[]
  /** Called after every activist has been reassigned. */
  onAssigned: () => void
}

export function BulkAssignDialog({
  open,
  onOpenChange,
  activistIds,
  onAssigned,
}: BulkAssignDialogProps) {
  const queryClient = useQueryClient()
  const [assignedTo, setAssignedTo] = useState<string>('')

  // Start each visit at an empty selection so a stale pick from last time
  // can't be submitted with one click. Adjusted during render rather than in
  // an effect so the old pick is never shown in the reopened dialog.
  const [wasOpen, setWasOpen] = useState(open)
  if (open !== wasOpen) {
    setWasOpen(open)
    if (open) setAssignedTo('')
  }

  const usersQuery = useQuery({
    queryKey: [API_PATH.USERS_ASSIGNABLE],
    queryFn: ({ signal }) => apiClient.getAssignableUsers(signal),
    enabled: open,
  })

  const mutation = useMutation({
    mutationFn: (userId: number) =>
      apiClient.assignActivists(activistIds, userId),
    onSuccess: ({ assigned }) => {
      const assigneeName =
        assignedTo === String(UNASSIGNED_USER_ID)
          ? 'Unassigned'
          : (usersQuery.data?.find((u) => String(u.id) === assignedTo)?.name ??
            'the selected user')
      toast.success(
        assigneeName === 'Unassigned'
          ? `Unassigned ${assigned} activist${assigned === 1 ? '' : 's'}`
          : `Assigned ${assigned} activist${assigned === 1 ? '' : 's'} to ${assigneeName}`,
      )
      queryClient.invalidateQueries({ queryKey: [API_PATH.ACTIVISTS_SEARCH] })
      queryClient.invalidateQueries({ queryKey: [API_PATH.ACTIVIST_GET] })
      onOpenChange(false)
      onAssigned()
    },
    onError: (err: Error) => {
      toast.error(err.message || 'Failed to assign activists')
    },
  })

  const handleOpenChange = (next: boolean) => {
    if (mutation.isPending) return
    onOpenChange(next)
  }

  const count = activistIds.length
  const tooMany = count > MAX_BULK_ASSIGN

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Assign activists</DialogTitle>
          <DialogDescription>
            {count} selected activist{count === 1 ? '' : 's'} will be assigned
            to the chosen user.
          </DialogDescription>
        </DialogHeader>

        <div className="flex flex-col gap-2">
          <Label htmlFor="bulk-assign-user">Assign to</Label>
          <Select
            value={assignedTo}
            onValueChange={setAssignedTo}
            disabled={
              mutation.isPending || usersQuery.isLoading || usersQuery.isError
            }
          >
            <SelectTrigger id="bulk-assign-user">
              <SelectValue placeholder="Select a user" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value={String(UNASSIGNED_USER_ID)}>
                Unassigned
              </SelectItem>
              {usersQuery.data?.map((user) => (
                <SelectItem key={user.id} value={String(user.id)}>
                  {user.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          {usersQuery.isError && (
            <p className="text-sm text-destructive">Failed to load users</p>
          )}
          {tooMany && (
            <p className="text-sm text-destructive">
              At most {MAX_BULK_ASSIGN} activists can be assigned at once.
            </p>
          )}
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
            onClick={() => mutation.mutate(parseInt(assignedTo, 10))}
            disabled={
              mutation.isPending || assignedTo === '' || count === 0 || tooMany
            }
          >
            {mutation.isPending ? 'Assigning...' : 'Assign'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
