'use client'

import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { ChevronDown, ChevronRight } from 'lucide-react'
import toast from 'react-hot-toast'
import { API_PATH, apiClient } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Collapse } from '@/components/ui/collapse'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'

// Free-text columns on the server (varchar(16) / varchar(32)), so these lists
// are the UI's own vocabulary rather than a database enum. Keep values within
// those lengths.
const INTERACTION_METHODS = [
  'Signal',
  'SMS',
  'Call',
  'Email',
  'WhatsApp',
  'Instagram',
  'Facebook',
  'X/Twitter',
  'Threads',
  'Bluesky',
  'Other',
] as const

const INTERACTION_OUTCOMES = [
  'Left message',
  'Had conversation',
  'No answer',
  'Could not contact',
] as const

// Radix Select disallows an empty item value, so outcome's "no answer given"
// choice needs a sentinel.
const OUTCOME_NONE = '__none__'

const DEFAULT_FOLLOWUP_DAYS = 3

// Organizers may log runs of interactions the same way (a phone bank is all
// Calls, a text bank all SMS), so the method carries over to the next dialog.
const METHOD_STORAGE_KEY = 'adb.logInteraction.method'

function readRememberedMethod(): string {
  try {
    // Session storage rather than local so as to persist only for the user's
    // current task.
    const stored = window.sessionStorage.getItem(METHOD_STORAGE_KEY)
    return INTERACTION_METHODS.some((m) => m === stored) ? stored! : ''
  } catch {
    // Storage can be unavailable (private mode)
    return ''
  }
}

function rememberMethod(method: string): void {
  try {
    window.sessionStorage.setItem(METHOD_STORAGE_KEY, method)
  } catch {
    // Not remembering the method is not worth failing the save over.
  }
}

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  activistId: number
  activistName: string
}

export function LogInteractionDialog({
  open,
  onOpenChange,
  activistId,
  activistName,
}: Props) {
  const queryClient = useQueryClient()

  const [method, setMethod] = useState(readRememberedMethod)
  const [outcome, setOutcome] = useState('')
  const [notes, setNotes] = useState('')
  const [assignSelf, setAssignSelf] = useState(true)
  const [followupExpanded, setFollowupExpanded] = useState(false)
  const [resetFollowup, setResetFollowup] = useState(false)
  const [setFollowup, setSetFollowup] = useState(false)
  // Kept as the raw string so the field can be emptied while retyping; a
  // number here would snap an empty box back to a digit and force the user to
  // type around it.
  const [followupDays, setFollowupDays] = useState(
    String(DEFAULT_FOLLOWUP_DAYS),
  )

  // Called on every close, so the next open starts blank rather than showing
  // the last activist's notes. The dialog's content is unmounted while closed,
  // so this never flashes on screen. The method is re-seeded from what was
  // last saved this session instead of being cleared.
  const resetForm = () => {
    setMethod(readRememberedMethod())
    setOutcome('')
    setNotes('')
    setAssignSelf(true)
    setFollowupExpanded(false)
    setResetFollowup(false)
    setSetFollowup(false)
    setFollowupDays(String(DEFAULT_FOLLOWUP_DAYS))
  }

  const followupDaysValue = Number(followupDays)
  const isFollowupDaysValid =
    Number.isInteger(followupDaysValue) &&
    followupDaysValue >= 1 &&
    followupDaysValue <= 365
  const canSave = !!method && (!setFollowup || isFollowupDaysValid)

  const mutation = useMutation({
    mutationFn: () =>
      apiClient.saveInteraction({
        activist_id: activistId,
        method,
        outcome: outcome === OUTCOME_NONE ? '' : outcome,
        notes: notes.trim(),
        assign_self: assignSelf,
        reset_followup: resetFollowup,
        set_followup: setFollowup,
        followup_days: isFollowupDaysValid ? followupDaysValue : 0,
      }),
    onSuccess: () => {
      rememberMethod(method)
      toast.success(`Interaction logged for ${activistName}`)
      // The interaction shows up in the timeline, and assigning to self or
      // touching the follow-up date changes the activist record itself.
      queryClient.invalidateQueries({
        queryKey: [API_PATH.ACTIVIST_TIMELINE, activistId],
      })
      queryClient.invalidateQueries({
        queryKey: [API_PATH.ACTIVIST_GET, activistId],
      })
      resetForm()
      onOpenChange(false)
    },
    onError: (err: Error) => {
      toast.error(err.message || 'Failed to log interaction')
    },
  })

  const handleOpenChange = (next: boolean) => {
    if (mutation.isPending) return
    if (!next) resetForm()
    onOpenChange(next)
  }

  // Collapsing the section also clears it, so "collapsed" always means "this
  // interaction changes nothing about the follow-up date" — otherwise a choice
  // made and then hidden would still be submitted with nothing on screen
  // saying so.
  const toggleFollowup = () => {
    setFollowupExpanded((expanded) => {
      if (expanded) {
        setResetFollowup(false)
        setSetFollowup(false)
        setFollowupDays(String(DEFAULT_FOLLOWUP_DAYS))
      }
      return !expanded
    })
  }

  const handleSubmit = () => {
    if (!method) {
      toast.error('Choose how you contacted them')
      return
    }
    if (setFollowup && !isFollowupDaysValid) {
      toast.error('Follow up in must be a whole number of days from 1 to 365')
      return
    }
    mutation.mutate()
  }

  const FollowupChevron = followupExpanded ? ChevronDown : ChevronRight

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      {/* The form is taller than a short phone screen, so the dialog is capped
          at the viewport and scrolls its own body rather than overflowing off
          the top and bottom where nothing can reach it. dvh (not vh) so mobile
          browser chrome sliding in and out doesn't clip the footer. */}
      <DialogContent className="flex max-h-[calc(100dvh-2rem)] flex-col">
        <DialogHeader className="shrink-0">
          <DialogTitle>Log interaction</DialogTitle>
          <DialogDescription>
            Record a conversation or contact attempt with {activistName}.
          </DialogDescription>
        </DialogHeader>

        {/* -mx-6/px-6 cancels the dialog's own padding so the scrollbar sits at
            its edge; py-1/-my-1 gives focus rings room without adding a gap.
            min-h-0 lets this shrink below its content so it, not the dialog,
            is what scrolls. */}
        <div className="-mx-6 -my-1 flex min-h-0 flex-1 flex-col gap-4 overflow-y-auto px-6 py-1">
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="interaction-method">Method</Label>
            <Select value={method} onValueChange={setMethod}>
              <SelectTrigger id="interaction-method">
                <SelectValue placeholder="Choose one" />
              </SelectTrigger>
              <SelectContent>
                {INTERACTION_METHODS.map((option) => (
                  <SelectItem key={option} value={option}>
                    {option}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="flex flex-col gap-1.5">
            <Label htmlFor="interaction-outcome">Outcome</Label>
            <Select value={outcome} onValueChange={setOutcome}>
              <SelectTrigger id="interaction-outcome">
                <SelectValue placeholder="Choose one" />
              </SelectTrigger>
              <SelectContent>
                {INTERACTION_OUTCOMES.map((option) => (
                  <SelectItem key={option} value={option}>
                    {option}
                  </SelectItem>
                ))}
                <SelectItem value={OUTCOME_NONE}>—</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="flex flex-col gap-1.5">
            <Label htmlFor="interaction-notes">Notes</Label>
            <Textarea
              id="interaction-notes"
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              rows={4}
            />
          </div>

          <div className="flex items-center gap-2">
            <Checkbox
              id="interaction-assign-self"
              checked={assignSelf}
              onCheckedChange={(checked) => setAssignSelf(Boolean(checked))}
            />
            <Label htmlFor="interaction-assign-self" className="cursor-pointer">
              Assign this activist to me
            </Label>
          </div>

          <div className="rounded-md border">
            <button
              type="button"
              onClick={toggleFollowup}
              aria-expanded={followupExpanded}
              className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm font-medium transition-colors hover:bg-accent"
            >
              <FollowupChevron
                className="h-4 w-4 shrink-0 text-muted-foreground"
                aria-hidden="true"
              />
              Follow-up
            </button>
            <Collapse open={followupExpanded}>
              <div className="flex flex-col gap-3 border-t p-3">
                <div className="flex items-center gap-2">
                  <Checkbox
                    id="interaction-reset-followup"
                    checked={resetFollowup}
                    disabled={setFollowup}
                    onCheckedChange={(checked) =>
                      setResetFollowup(Boolean(checked))
                    }
                  />
                  <Label
                    htmlFor="interaction-reset-followup"
                    className="cursor-pointer"
                  >
                    Clear follow-up date
                  </Label>
                </div>

                <div className="flex flex-wrap items-center gap-2">
                  <Checkbox
                    id="interaction-set-followup"
                    checked={setFollowup}
                    disabled={resetFollowup}
                    onCheckedChange={(checked) =>
                      setSetFollowup(Boolean(checked))
                    }
                  />
                  <Label
                    htmlFor="interaction-set-followup"
                    className="cursor-pointer"
                  >
                    Follow up in
                  </Label>
                  <Input
                    type="number"
                    aria-label="Follow-up days"
                    className="w-20"
                    min={1}
                    max={365}
                    value={followupDays}
                    disabled={!setFollowup}
                    aria-invalid={setFollowup && !isFollowupDaysValid}
                    onChange={(e) => setFollowupDays(e.target.value)}
                  />
                  <span className="text-sm text-muted-foreground">days</span>
                </div>
              </div>
            </Collapse>
          </div>
        </div>

        <DialogFooter className="shrink-0">
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
            onClick={handleSubmit}
            disabled={!canSave || mutation.isPending}
          >
            {mutation.isPending ? 'Saving...' : 'Save'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
