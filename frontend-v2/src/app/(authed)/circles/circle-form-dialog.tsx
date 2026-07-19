'use client'

import { useMemo } from 'react'
import { useForm } from '@tanstack/react-form'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import toast from 'react-hot-toast'
import { z } from 'zod'
import {
  API_PATH,
  apiClient,
  CircleGroup,
  SaveCircleMemberParams,
} from '@/lib/api'
import { findPointPerson } from '@/lib/members'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { TagInput } from '@/components/tag-input'
import type { CircleMode } from './search-params'

// Trimming matches the legacy Vue form's `v-model.trim`.
const circleFormSchema = z.object({
  name: z.string().trim().min(1, 'Circle name must not be blank'),
  description: z.string().trim(),
  meetingTime: z.string().trim(),
  meetingLocation: z.string().trim(),
  coords: z.string().trim(),
  visible: z.boolean(),
  host: z.array(z.string()).max(1),
  members: z.array(z.string()),
})

type CircleFormValues = z.input<typeof circleFormSchema>

function buildInitialValues(circle: CircleGroup | null): CircleFormValues {
  if (!circle) {
    return {
      name: '',
      description: '',
      meetingTime: '',
      meetingLocation: '',
      coords: '',
      visible: false,
      host: [],
      members: [],
    }
  }

  const pointPerson = findPointPerson(circle.members)
  const host = pointPerson ? [pointPerson.name] : []
  const members = circle.members
    .filter((m) => !m.point_person)
    .map((m) => m.name)

  return {
    name: circle.name,
    description: circle.description,
    meetingTime: circle.meeting_time,
    meetingLocation: circle.meeting_location,
    coords: circle.coords,
    visible: circle.visible,
    host,
    members,
  }
}

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  mode: CircleMode
  circle: CircleGroup | null
  activistNames: string[]
}

export function CircleFormDialog({
  open,
  onOpenChange,
  mode,
  circle,
  activistNames,
}: Props) {
  const queryClient = useQueryClient()
  const isGeo = mode === 'geo'
  const isEditing = circle !== null

  const mutation = useMutation({
    mutationFn: (value: z.output<typeof circleFormSchema>) => {
      // The host wins over a duplicate member entry (legacy parity).
      const hostName = value.host[0]
      const notHost = (name: string) => name !== hostName
      const memberParams: SaveCircleMemberParams[] = [
        ...(hostName ? [{ name: hostName, point_person: true }] : []),
        ...value.members
          .filter(notHost)
          .map((name) => ({ name, point_person: false })),
      ]

      return apiClient.saveCircle({
        id: circle?.id ?? 0,
        name: value.name,
        type: isGeo ? 'geo-circle' : 'circle',
        description: value.description,
        meeting_time: value.meetingTime,
        meeting_location: value.meetingLocation,
        coords: value.coords,
        visible: value.visible,
        members: memberParams,
      })
    },
    onSuccess: (savedCircle) => {
      queryClient.invalidateQueries({ queryKey: [API_PATH.CIRCLE_LIST] })
      toast.success(`${savedCircle.name} saved`)
      onOpenChange(false)
    },
    onError: (err: Error) => {
      toast.error(err.message || 'Failed to save circle. Please try again.')
    },
  })

  const initialValues = useMemo(() => buildInitialValues(circle), [circle])

  const form = useForm({
    defaultValues: initialValues,
    onSubmit: async ({ value }) => {
      const parsed = circleFormSchema.safeParse(value)
      if (!parsed.success) {
        toast.error(
          parsed.error.issues[0]?.message ?? 'Please check the form fields',
        )
        return
      }
      await mutation.mutateAsync(parsed.data)
    },
  })

  const isSubmitting = mutation.isPending

  const handleOpenChange = (next: boolean) => {
    if (isSubmitting) return
    onOpenChange(next)
  }

  const circleLabel = isGeo ? 'Geo-Circle' : 'Circle'

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-md">
        <form
          onSubmit={(e) => {
            e.preventDefault()
            form.handleSubmit()
          }}
          className="flex flex-col gap-4"
        >
          <DialogHeader>
            <DialogTitle>
              {isEditing ? 'Edit' : 'New'} {circleLabel}
            </DialogTitle>
          </DialogHeader>

          <form.Field name="name">
            {(field) => (
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="circle-name">Name</Label>
                <Input
                  id="circle-name"
                  value={field.state.value}
                  onChange={(e) => field.handleChange(e.target.value)}
                  onBlur={field.handleBlur}
                  required
                  autoFocus
                />
              </div>
            )}
          </form.Field>

          <form.Field name="description">
            {(field) => (
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="circle-description">
                  Description{isGeo ? ' or Notes' : ''}
                </Label>
                <Input
                  id="circle-description"
                  value={field.state.value}
                  onChange={(e) => field.handleChange(e.target.value)}
                />
              </div>
            )}
          </form.Field>

          {!isGeo && (
            <>
              <form.Field name="meetingTime">
                {(field) => (
                  <div className="flex flex-col gap-1.5">
                    <Label htmlFor="circle-meeting-time">
                      Meeting Day &amp; Time
                    </Label>
                    <Input
                      id="circle-meeting-time"
                      value={field.state.value}
                      onChange={(e) => field.handleChange(e.target.value)}
                    />
                  </div>
                )}
              </form.Field>

              <form.Field name="meetingLocation">
                {(field) => (
                  <div className="flex flex-col gap-1.5">
                    <Label htmlFor="circle-meeting-location">
                      Meeting Location
                    </Label>
                    <Input
                      id="circle-meeting-location"
                      value={field.state.value}
                      onChange={(e) => field.handleChange(e.target.value)}
                    />
                  </div>
                )}
              </form.Field>
            </>
          )}

          <form.Field name="coords">
            {(field) => (
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="circle-coords">Coordinates</Label>
                <Input
                  id="circle-coords"
                  value={field.state.value}
                  onChange={(e) => field.handleChange(e.target.value)}
                  disabled={isGeo}
                />
                {isGeo && (
                  <p className="text-xs text-muted-foreground">
                    Coordinates for geo-circles are calculated automatically
                    from member locations.
                  </p>
                )}
              </div>
            )}
          </form.Field>

          <form.Field name="visible">
            {(field) => (
              <div className="flex items-center gap-2">
                <Checkbox
                  id="circle-visible"
                  checked={field.state.value}
                  onCheckedChange={(checked) =>
                    field.handleChange(Boolean(checked))
                  }
                />
                <Label htmlFor="circle-visible">Visible to public</Label>
              </div>
            )}
          </form.Field>

          <form.Field name="host">
            {(field) => (
              <TagInput
                label="Host"
                value={field.state.value}
                onChange={field.handleChange}
                options={activistNames}
                placeholder="Search by name..."
                single
              />
            )}
          </form.Field>

          {isGeo && (
            <form.Field name="members">
              {(field) => (
                <TagInput
                  label="Members"
                  value={field.state.value}
                  onChange={field.handleChange}
                  options={activistNames}
                  placeholder="Search by name..."
                />
              )}
            </form.Field>
          )}

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => handleOpenChange(false)}
              disabled={isSubmitting}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? 'Saving...' : 'Save'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
