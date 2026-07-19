'use client'

import { useMemo } from 'react'
import { useForm } from '@tanstack/react-form'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import toast from 'react-hot-toast'
import { Loader2, Save } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Checkbox } from '@/components/ui/checkbox'
import {
  API_PATH,
  apiClient,
  WorkingGroup,
  WorkingGroupMemberInput,
  WorkingGroupSavePayload,
} from '@/lib/api'
import { TagInput } from '@/components/tag-input'
import { findPointPerson } from '@/lib/members'

interface WorkingGroupFormValues {
  name: string
  email: string
  description: string
  meetingTime: string
  meetingLocation: string
  visible: boolean
  pointPerson: string[]
  members: string[]
  nonMembers: string[]
}

function buildInitialValues(
  workingGroup: WorkingGroup | null,
): WorkingGroupFormValues {
  if (!workingGroup) {
    return {
      name: '',
      email: '',
      description: '',
      meetingTime: '',
      meetingLocation: '',
      visible: false,
      pointPerson: [],
      members: [],
      nonMembers: [],
    }
  }

  const pointPersonName = findPointPerson(workingGroup.members)?.name
  const pointPerson = pointPersonName ? [pointPersonName] : []
  const members = workingGroup.members
    .filter((m) => !m.point_person && !m.non_member_on_mailing_list)
    .map((m) => m.name)
  const nonMembers = workingGroup.members
    .filter((m) => m.non_member_on_mailing_list)
    .map((m) => m.name)

  return {
    name: workingGroup.name,
    email: workingGroup.email,
    description: workingGroup.description,
    meetingTime: workingGroup.meeting_time,
    meetingLocation: workingGroup.meeting_location,
    visible: workingGroup.visible,
    pointPerson,
    members,
    nonMembers,
  }
}

export function WorkingGroupFormDialog({
  workingGroup,
  organizerNames,
  activistNames,
  onClose,
}: {
  /** Null when creating a new working group. */
  workingGroup: WorkingGroup | null
  organizerNames: string[]
  activistNames: string[]
  onClose: () => void
}) {
  const queryClient = useQueryClient()

  const mutation = useMutation({
    mutationFn: (payload: WorkingGroupSavePayload) =>
      apiClient.saveWorkingGroup(payload),
    onSuccess: (saved) => {
      queryClient.invalidateQueries({
        queryKey: [API_PATH.WORKING_GROUP_LIST],
      })
      toast.success(`${saved.name} saved`)
      onClose()
    },
    onError: (err: unknown) => {
      const message =
        err instanceof Error ? err.message : 'Failed to save working group'
      toast.error(`Error: ${message}`)
    },
  })

  const initialValues = useMemo(
    () => buildInitialValues(workingGroup),
    [workingGroup],
  )

  const form = useForm({
    defaultValues: initialValues,
    onSubmit: async ({ value }) => {
      const name = value.name.trim()
      if (!name) {
        toast.error('Working group name must not be blank')
        return
      }
      const email = value.email.trim()
      if (!email.includes('@')) {
        toast.error(`Working group email must contain @: ${email}`)
        return
      }

      // The point person wins over duplicate member entries (legacy parity).
      const pointPersonName = value.pointPerson[0]
      const members: WorkingGroupMemberInput[] = []
      if (pointPersonName) {
        members.push({
          name: pointPersonName,
          point_person: true,
          non_member_on_mailing_list: false,
        })
      }
      for (const m of value.members) {
        if (m === pointPersonName) continue
        members.push({
          name: m,
          point_person: false,
          non_member_on_mailing_list: false,
        })
      }
      for (const m of value.nonMembers) {
        if (m === pointPersonName) continue
        members.push({
          name: m,
          point_person: false,
          non_member_on_mailing_list: true,
        })
      }

      await mutation.mutateAsync({
        id: workingGroup?.id,
        name,
        email,
        visible: value.visible,
        description: value.description,
        meeting_time: value.meetingTime,
        meeting_location: value.meetingLocation,
        // Not editable here (nor in legacy); round-trip so saving doesn't clear it.
        coords: workingGroup?.coords ?? '',
        members,
      })
    },
  })

  const isSubmitting = mutation.isPending

  return (
    <Dialog open onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>
            {workingGroup ? 'Edit' : 'New'} Working Group
          </DialogTitle>
        </DialogHeader>

        <form
          onSubmit={(e) => {
            e.preventDefault()
            form.handleSubmit()
          }}
          className="space-y-4"
        >
          <form.Field name="name">
            {(field) => (
              <div className="space-y-1">
                <Label htmlFor="wg-name">Name</Label>
                <Input
                  id="wg-name"
                  value={field.state.value}
                  onChange={(e) => field.handleChange(e.target.value)}
                  onBlur={field.handleBlur}
                  required
                />
              </div>
            )}
          </form.Field>

          <form.Field name="email">
            {(field) => (
              <div className="space-y-1">
                <Label htmlFor="wg-email">Email</Label>
                <Input
                  id="wg-email"
                  type="email"
                  value={field.state.value}
                  onChange={(e) => field.handleChange(e.target.value)}
                  onBlur={field.handleBlur}
                  required
                />
              </div>
            )}
          </form.Field>

          <form.Field name="description">
            {(field) => (
              <div className="space-y-1">
                <Label htmlFor="wg-description">Description</Label>
                <Input
                  id="wg-description"
                  value={field.state.value}
                  onChange={(e) => field.handleChange(e.target.value)}
                />
              </div>
            )}
          </form.Field>

          <form.Field name="meetingTime">
            {(field) => (
              <div className="space-y-1">
                <Label htmlFor="wg-meeting-time">Meeting Day &amp; Time</Label>
                <Input
                  id="wg-meeting-time"
                  value={field.state.value}
                  onChange={(e) => field.handleChange(e.target.value)}
                />
              </div>
            )}
          </form.Field>

          <form.Field name="meetingLocation">
            {(field) => (
              <div className="space-y-1">
                <Label htmlFor="wg-meeting-location">Meeting Location</Label>
                <Input
                  id="wg-meeting-location"
                  value={field.state.value}
                  onChange={(e) => field.handleChange(e.target.value)}
                />
              </div>
            )}
          </form.Field>

          <form.Field name="visible">
            {(field) => (
              <label className="flex items-center gap-2 text-sm">
                <Checkbox
                  checked={field.state.value}
                  onCheckedChange={(checked) =>
                    field.handleChange(Boolean(checked))
                  }
                />
                Visible
              </label>
            )}
          </form.Field>

          <form.Field name="pointPerson">
            {(field) => (
              <TagInput
                label="Point Person"
                options={organizerNames}
                value={field.state.value}
                onChange={field.handleChange}
                max={1}
              />
            )}
          </form.Field>

          <form.Field name="members">
            {(field) => (
              <TagInput
                label="Members"
                options={organizerNames}
                value={field.state.value}
                onChange={field.handleChange}
              />
            )}
          </form.Field>

          <form.Field name="nonMembers">
            {(field) => (
              <TagInput
                label="Non-members on Mailing List"
                options={activistNames}
                value={field.state.value}
                onChange={field.handleChange}
              />
            )}
          </form.Field>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={onClose}
              disabled={isSubmitting}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <Save className="h-4 w-4" />
              )}
              Save
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
