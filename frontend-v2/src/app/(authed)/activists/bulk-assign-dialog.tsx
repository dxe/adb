'use client'

import { useState } from 'react'
import {
  useMutation,
  useQuery,
  useQueryClient,
  type InfiniteData,
  type QueryClient,
} from '@tanstack/react-query'
import toast from 'react-hot-toast'
import { API_PATH, apiClient, type QueryActivistResult } from '@/lib/api'
import { activistKeys } from '@/lib/query-keys'
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

/**
 * Writes the new assignee into every cached activist list so that new values
 * will appear in the table. Related queries are invalidated so they can be
 * manually refreshed later when the user is ready.
 *
 * After a bulk operation, the user may want to perform another bulk operation
 * on the same selection. Refetching would disorient the user and lose their
 * selection: rows that no longer match the filters would disappear, and rows
 * would be reordered if sorting is enabled on the affected column.
 */
function updateCachedAssignee(
  queryClient: QueryClient,
  activistIds: number[],
  assignedTo: number,
  assignedToName: string,
) {
  const ids = new Set(activistIds)
  queryClient.setQueriesData<InfiniteData<QueryActivistResult>>(
    { queryKey: activistKeys.lists() },
    (data) =>
      data && {
        ...data,
        pages: data.pages.map((page) =>
          page.activists.some((activist) => ids.has(activist.id))
            ? {
                ...page,
                activists: page.activists.map((activist) =>
                  ids.has(activist.id)
                    ? {
                        ...activist,
                        assigned_to: assignedTo,
                        assigned_to_name: assignedToName,
                      }
                    : activist,
                ),
              }
            : page,
        ),
      },
  )
  queryClient.invalidateQueries({
    queryKey: activistKeys.lists(),
    refetchType: 'none',
  })
  queryClient.invalidateQueries({
    queryKey: activistKeys.counts(),
    refetchType: 'none',
  })
  queryClient.invalidateQueries({ queryKey: activistKeys.details() })
}

interface BulkAssignDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  activistIds: number[]
  /** Called with the new assignee once the assignment succeeds. */
  onAssigned?: (assigneeId: number) => void
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
    onSuccess: ({ assigned }, userId) => {
      // Empty for "Unassigned", which no user matches.
      const assigneeName =
        usersQuery.data?.find((u) => String(u.id) === assignedTo)?.name ?? ''
      toast.success(
        userId === UNASSIGNED_USER_ID
          ? `Unassigned ${assigned} activist${assigned === 1 ? '' : 's'}`
          : `Assigned ${assigned} activist${assigned === 1 ? '' : 's'} to ${assigneeName || 'the selected user'}`,
      )
      updateCachedAssignee(queryClient, activistIds, userId, assigneeName)
      onAssigned?.(userId)
      onOpenChange(false)
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
