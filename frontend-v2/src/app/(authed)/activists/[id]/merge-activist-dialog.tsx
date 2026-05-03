'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import toast from 'react-hot-toast'
import { API_PATH, apiClient } from '@/lib/api'
import { useAuthedPageContext } from '@/hooks/useAuthedPageContext'
import { useActivistRegistry } from '../../events/useActivistRegistry'
import { SuggestionInput } from '../../events/suggestion-input'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
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

export function MergeActivistDialog({
  open,
  onOpenChange,
  activistId,
  activistName,
}: Props) {
  const router = useRouter()
  const queryClient = useQueryClient()
  const { user } = useAuthedPageContext()
  const { registry } = useActivistRegistry(user.ChapterID)

  // Raw text in the input as the user types.
  const [inputValue, setInputValue] = useState('')
  // Name of an activist explicitly selected from the suggestion list.
  const [selectedValue, setSelectedValue] = useState('')

  const mutation = useMutation({
    mutationFn: (targetName: string) =>
      apiClient.mergeActivist(activistId, targetName),
    onSuccess: (_data, targetName) => {
      toast.success(`${activistName} was merged into ${targetName}`)
      queryClient.invalidateQueries({ queryKey: [API_PATH.ACTIVISTS_SEARCH] })
      queryClient.invalidateQueries({
        queryKey: [API_PATH.ACTIVIST_LIST_BASIC],
      })
      onOpenChange(false)
      router.push('/activists')
    },
    onError: (err: Error) => {
      toast.error(err.message || 'Failed to merge activist')
    },
  })

  const handleOpenChange = (next: boolean) => {
    if (mutation.isPending) return
    if (!next) {
      setInputValue('')
      setSelectedValue('')
    }
    onOpenChange(next)
  }

  const getSuggestions = (input: string) =>
    registry.getSuggestions(input).filter((name) => name !== activistName)

  const isTargetValid =
    selectedValue.trim().length > 0 &&
    // Cannot merge activist with self. (Names are unique within chapter, and
    // backend endpoint will not allow merging across chapters.)
    selectedValue !== activistName &&
    registry.getActivist(selectedValue) !== null

  const handleSubmit = () => {
    if (!isTargetValid) {
      toast.error('Choose an activist to merge into')
      return
    }
    mutation.mutate(selectedValue)
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Merge activist</DialogTitle>
        </DialogHeader>
        <div className="text-sm space-y-2">
          <p>
            Merging activists is used to combine redundant activist entries.
          </p>
          <p>Merging this activist does the following:</p>
          <ul className="list-disc pl-5 space-y-1">
            <li>
              All of {activistName}&apos;s event attendance will be moved to the
              target activist.
            </li>
            <li>
              The target&apos;s name will be kept &mdash; {activistName}&apos;s
              name will not replace it.
            </li>
            <li>
              {activistName}&apos;s email, phone, and address / location will
              replace the corresponding fields in the target{' '}
              <em>if updated more recently</em>.
            </li>
            <li>
              {activistName}&apos;s pronouns, language, notes, and most other
              fields will each replace the corresponding field in the target{' '}
              <em>only if the target&apos;s value is blank</em>.
            </li>
            <li>
              Yes / no flags (e.g. MPI, hiatus, voting agreement) will be set to
              &ldquo;yes&rdquo; <em>if either activist has them set</em>.
            </li>
            <li>
              The target&apos;s activist level will be set to the{' '}
              <em>higher of the two activists&apos; levels</em>.
            </li>
            <li>{activistName} will be hidden.</li>
          </ul>
          <p className="font-semibold pt-2">
            Merge {activistName} into another activist:
          </p>
        </div>
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="merge-target">Target activist</Label>
          <SuggestionInput
            id="merge-target"
            value={inputValue}
            onValueChange={(v) => {
              setInputValue(v)
              setSelectedValue('')
            }}
            getSuggestions={getSuggestions}
            onCommit={(meta) => {
              if (meta.fromSuggestion) {
                setSelectedValue(meta.value)
              }
            }}
            placeholder="Start typing a name..."
          />
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
            variant="destructive"
            onClick={handleSubmit}
            disabled={!isTargetValid || mutation.isPending}
          >
            {mutation.isPending ? 'Merging...' : 'Merge activist'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
