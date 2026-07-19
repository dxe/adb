'use client'

import { FormEvent, useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import toast from 'react-hot-toast'
import {
  API_PATH,
  apiClient,
  CircleGroup,
  SaveCircleMemberParams,
} from '@/lib/api'
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

  const [name, setName] = useState(circle?.name ?? '')
  const [description, setDescription] = useState(circle?.description ?? '')
  const [meetingTime, setMeetingTime] = useState(circle?.meeting_time ?? '')
  const [meetingLocation, setMeetingLocation] = useState(
    circle?.meeting_location ?? '',
  )
  const [coords, setCoords] = useState(circle?.coords ?? '')
  const [visible, setVisible] = useState(circle?.visible ?? false)
  const [host, setHost] = useState<string[]>(() => {
    const pointPerson = circle?.members.find((m) => m.point_person)
    return pointPerson ? [pointPerson.name] : []
  })
  const [members, setMembers] = useState<string[]>(
    () =>
      circle?.members.filter((m) => !m.point_person).map((m) => m.name) ?? [],
  )

  const mutation = useMutation({
    mutationFn: () => {
      const memberParams: SaveCircleMemberParams[] = []
      if (host.length > 0) {
        memberParams.push({ name: host[0], point_person: true })
      }
      members.forEach((memberName) => {
        const sameAsHost = host.length > 0 && host[0] === memberName
        if (!sameAsHost) {
          memberParams.push({ name: memberName, point_person: false })
        }
      })

      // The legacy Vue form used `v-model.trim` on these text fields, so trim
      // them at submit time for parity.
      return apiClient.saveCircle({
        id: circle?.id ?? 0,
        name: name.trim(),
        type: isGeo ? 'geo-circle' : 'circle',
        description: description.trim(),
        meeting_time: meetingTime.trim(),
        meeting_location: meetingLocation.trim(),
        coords: coords.trim(),
        visible,
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

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault()
    if (!name.trim()) {
      toast.error('Circle name must not be blank')
      return
    }
    mutation.mutate()
  }

  const handleOpenChange = (next: boolean) => {
    if (mutation.isPending) return
    onOpenChange(next)
  }

  const circleLabel = isGeo ? 'Geo-Circle' : 'Circle'

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-md">
        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <DialogHeader>
            <DialogTitle>
              {isEditing ? 'Edit' : 'New'} {circleLabel}
            </DialogTitle>
          </DialogHeader>

          <div className="flex flex-col gap-1.5">
            <Label htmlFor="circle-name">Name</Label>
            <Input
              id="circle-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              autoFocus
            />
          </div>

          <div className="flex flex-col gap-1.5">
            <Label htmlFor="circle-description">
              Description{isGeo ? ' or Notes' : ''}
            </Label>
            <Input
              id="circle-description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
          </div>

          {!isGeo && (
            <>
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="circle-meeting-time">
                  Meeting Day &amp; Time
                </Label>
                <Input
                  id="circle-meeting-time"
                  value={meetingTime}
                  onChange={(e) => setMeetingTime(e.target.value)}
                />
              </div>

              <div className="flex flex-col gap-1.5">
                <Label htmlFor="circle-meeting-location">
                  Meeting Location
                </Label>
                <Input
                  id="circle-meeting-location"
                  value={meetingLocation}
                  onChange={(e) => setMeetingLocation(e.target.value)}
                />
              </div>
            </>
          )}

          <div className="flex flex-col gap-1.5">
            <Label htmlFor="circle-coords">Coordinates</Label>
            <Input
              id="circle-coords"
              value={coords}
              onChange={(e) => setCoords(e.target.value)}
              disabled={isGeo}
            />
            {isGeo && (
              <p className="text-xs text-muted-foreground">
                Coordinates for geo-circles are calculated automatically from
                member locations.
              </p>
            )}
          </div>

          <div className="flex items-center gap-2">
            <Checkbox
              id="circle-visible"
              checked={visible}
              onCheckedChange={(value) => setVisible(Boolean(value))}
            />
            <Label htmlFor="circle-visible">Visible to public</Label>
          </div>

          <div className="flex flex-col gap-1.5">
            <Label htmlFor="circle-host">Host</Label>
            <TagInput
              id="circle-host"
              value={host}
              onChange={setHost}
              options={activistNames}
              placeholder="Search by name..."
              max={1}
            />
          </div>

          {isGeo && (
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="circle-members">Members</Label>
              <TagInput
                id="circle-members"
                value={members}
                onChange={setMembers}
                options={activistNames}
                placeholder="Search by name..."
              />
            </div>
          )}

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => handleOpenChange(false)}
              disabled={mutation.isPending}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={mutation.isPending}>
              {mutation.isPending ? 'Saving...' : 'Save'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
