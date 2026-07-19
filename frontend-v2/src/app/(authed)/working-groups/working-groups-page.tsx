'use client'

import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { API_PATH, apiClient, WorkingGroup } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Eye, EyeOff, Loader2, Plus } from 'lucide-react'
import { WorkingGroupTable } from './working-group-table'
import { WorkingGroupFormDialog } from './working-group-form-dialog'
import { DeleteWorkingGroupDialog } from './delete-working-group-dialog'

// Sentinel for the create dialog, which has no WorkingGroup to edit yet.
const NEW_WORKING_GROUP = 'new' as const

export default function WorkingGroupsPage() {
  const {
    data: workingGroups,
    isLoading: isWorkingGroupsLoading,
    isError,
  } = useQuery({
    queryKey: [API_PATH.WORKING_GROUP_LIST],
    queryFn: ({ signal }) => apiClient.getWorkingGroups(signal),
  })

  const {
    data: organizersResp,
    isLoading: isOrganizersLoading,
    isError: isOrganizersError,
  } = useQuery({
    queryKey: [API_PATH.ACTIVIST_NAMES_GET_ORGANIZERS],
    queryFn: ({ signal }) => apiClient.getOrganizerNames(signal),
  })

  const {
    data: activistsResp,
    isLoading: isActivistsLoading,
    isError: isActivistsError,
  } = useQuery({
    queryKey: [API_PATH.ACTIVIST_NAMES_GET],
    queryFn: ({ signal }) => apiClient.getActivistNames(signal),
  })

  const [membersVisible, setMembersVisible] = useState(false)
  const [editingTarget, setEditingTarget] = useState<
    WorkingGroup | typeof NEW_WORKING_GROUP | null
  >(null)
  const [deletingTarget, setDeletingTarget] = useState<WorkingGroup | null>(
    null,
  )

  const organizerNames = organizersResp?.activist_names ?? []
  const activistNames = activistsResp?.activist_names ?? []

  const isLoading =
    isWorkingGroupsLoading || isOrganizersLoading || isActivistsLoading

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-start justify-between gap-3">
        <div className="flex flex-col gap-1">
          <h1 className="text-2xl font-semibold">Working Groups</h1>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            onClick={() => setMembersVisible((v) => !v)}
          >
            {membersVisible ? (
              <EyeOff className="h-4 w-4" />
            ) : (
              <Eye className="h-4 w-4" />
            )}
            {membersVisible ? 'Hide' : 'Show'} members
          </Button>
          <Button onClick={() => setEditingTarget(NEW_WORKING_GROUP)}>
            <Plus className="h-4 w-4" />
            New Working Group
          </Button>
        </div>
      </div>

      {isLoading ? (
        <div className="flex items-center gap-2 text-muted-foreground text-sm">
          <Loader2 className="h-4 w-4 animate-spin" />
          Loading working groups...
        </div>
      ) : isError || !workingGroups ? (
        <div className="text-sm text-destructive">
          Failed to load working groups. Please try again.
        </div>
      ) : (
        <>
          {(isOrganizersError || isActivistsError) && (
            <div className="text-sm text-destructive">
              Failed to load activist names. The point person and member
              autocomplete will be unavailable — reload the page to try again.
            </div>
          )}
          <WorkingGroupTable
            workingGroups={workingGroups}
            membersVisible={membersVisible}
            onEdit={setEditingTarget}
            onDelete={setDeletingTarget}
          />
        </>
      )}

      {editingTarget !== null && (
        <WorkingGroupFormDialog
          workingGroup={
            editingTarget === NEW_WORKING_GROUP ? null : editingTarget
          }
          organizerNames={organizerNames}
          activistNames={activistNames}
          onClose={() => setEditingTarget(null)}
        />
      )}

      {deletingTarget && (
        <DeleteWorkingGroupDialog
          workingGroup={deletingTarget}
          onClose={() => setDeletingTarget(null)}
        />
      )}
    </div>
  )
}
