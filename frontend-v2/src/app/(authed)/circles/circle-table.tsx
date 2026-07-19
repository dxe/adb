'use client'

import { useMemo, useState } from 'react'
import { isAfter, isValid, parseISO, subDays } from 'date-fns'
import { Pencil, Trash2 } from 'lucide-react'
import type { CircleGroup } from '@/lib/api'
import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { SortIndicator } from '@/components/ui/sort-indicator'
import { cn } from '@/lib/utils'
import type { CircleMode } from './search-params'

type SortColumn = 'name' | 'last_meeting'
type SortState = { column: SortColumn; desc: boolean }

/** Mirrors the legacy `colorLastMeeting` freshness indicator: red (stale) if
 * the last event was more than 32 days ago, yellow between 15-32 days ago,
 * and green within the last 15 days. */
function lastMeetingTone(text: string): 'stale' | 'warning' | 'fresh' | null {
  if (!text) return null
  const time = parseISO(text)
  if (!isValid(time)) return null
  const now = new Date()
  let tone: 'stale' | 'warning' | 'fresh' = 'stale'
  if (isAfter(time, subDays(now, 32))) tone = 'warning'
  if (isAfter(time, subDays(now, 15))) tone = 'fresh'
  return tone
}

const toneClasses: Record<'stale' | 'warning' | 'fresh', string> = {
  stale: 'bg-red-100 text-red-700',
  warning: 'bg-amber-100 text-amber-800',
  fresh: 'bg-emerald-100 text-emerald-700',
}

function hostName(circle: CircleGroup): string {
  // There should only ever be one point person.
  return circle.members.find((m) => m.point_person)?.name ?? ''
}

function memberCount(circle: CircleGroup): number {
  return circle.members.filter((m) => !m.non_member_on_mailing_list).length
}

export function CircleTable({
  circles,
  mode,
  membersVisible,
  onEdit,
  onDelete,
}: {
  circles: CircleGroup[]
  mode: CircleMode
  membersVisible: boolean
  onEdit: (circle: CircleGroup) => void
  onDelete: (circle: CircleGroup) => void
}) {
  const [sort, setSort] = useState<SortState>({ column: 'name', desc: false })

  const sorted = useMemo(() => {
    const copy = [...circles]
    copy.sort((a, b) => {
      let cmp: number
      if (sort.column === 'name') {
        cmp = a.name.localeCompare(b.name)
      } else {
        cmp = a.last_meeting.localeCompare(b.last_meeting)
      }
      return sort.desc ? -cmp : cmp
    })
    return copy
  }, [circles, sort])

  const toggleSort = (column: SortColumn) => {
    setSort((prev) =>
      prev.column === column
        ? { column, desc: !prev.desc }
        : { column, desc: false },
    )
  }

  const sortIndicatorFor = (column: SortColumn) =>
    sort.column === column ? (sort.desc ? 'desc' : 'asc') : false

  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-[96px]">Actions</TableHead>
            <TableHead>
              <button
                type="button"
                onClick={() => toggleSort('name')}
                className="flex items-center gap-1"
              >
                <span>Name</span>
                <SortIndicator sorted={sortIndicatorFor('name')} />
              </button>
            </TableHead>
            <TableHead>Host</TableHead>
            {mode === 'interest' && (
              <TableHead>
                <button
                  type="button"
                  onClick={() => toggleSort('last_meeting')}
                  className="flex items-center gap-1"
                >
                  <span>Last Event</span>
                  <SortIndicator sorted={sortIndicatorFor('last_meeting')} />
                </button>
              </TableHead>
            )}
            {mode === 'geo' && !membersVisible && (
              <TableHead>Total Members</TableHead>
            )}
            {mode === 'geo' && membersVisible && <TableHead>Members</TableHead>}
          </TableRow>
        </TableHeader>
        <TableBody>
          {sorted.length ? (
            sorted.map((circle) => {
              const tone = lastMeetingTone(circle.last_meeting)
              return (
                <TableRow key={circle.id}>
                  <TableCell>
                    <div className="flex gap-1">
                      <Button
                        type="button"
                        variant="outline"
                        size="icon"
                        aria-label={`Edit ${circle.name}`}
                        onClick={() => onEdit(circle)}
                      >
                        <Pencil className="h-4 w-4 text-primary" />
                      </Button>
                      <Button
                        type="button"
                        variant="outline"
                        size="icon"
                        aria-label={`Delete ${circle.name}`}
                        onClick={() => onDelete(circle)}
                      >
                        <Trash2 className="h-4 w-4 text-destructive" />
                      </Button>
                    </div>
                  </TableCell>
                  <TableCell className="font-medium">{circle.name}</TableCell>
                  <TableCell>{hostName(circle)}</TableCell>
                  {mode === 'interest' && (
                    <TableCell>
                      <span
                        className={cn(
                          'rounded-full px-2 py-0.5 text-xs font-medium',
                          tone
                            ? toneClasses[tone]
                            : 'bg-muted text-muted-foreground',
                        )}
                      >
                        {circle.last_meeting || 'None'}
                      </span>
                    </TableCell>
                  )}
                  {mode === 'geo' && !membersVisible && (
                    <TableCell>{memberCount(circle)}</TableCell>
                  )}
                  {mode === 'geo' && membersVisible && (
                    <TableCell>
                      <ul className="list-disc space-y-0.5 pl-4">
                        {circle.members
                          .filter((m) => !m.point_person)
                          .map((m) => (
                            <li key={m.name}>{m.name}</li>
                          ))}
                      </ul>
                    </TableCell>
                  )}
                </TableRow>
              )
            })
          ) : (
            <TableRow>
              <TableCell
                colSpan={mode === 'interest' ? 4 : 4}
                className="py-6 text-center text-muted-foreground"
              >
                No {mode === 'interest' ? 'circles' : 'geo-circles'} found.
              </TableCell>
            </TableRow>
          )}
        </TableBody>
      </Table>
    </div>
  )
}
