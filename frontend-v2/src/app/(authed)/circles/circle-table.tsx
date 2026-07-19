'use client'

import { useMemo, useState } from 'react'
import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable,
} from '@tanstack/react-table'
import { isAfter, isValid, parseISO, subDays } from 'date-fns'
import { Pencil, Trash2 } from 'lucide-react'
import type { CircleGroup } from '@/lib/api'
import { countMailingListMembers, findPointPerson } from '@/lib/members'
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

const STALE_AFTER_DAYS = 32
const FRESH_WITHIN_DAYS = 15

// Mirrors the legacy `colorLastMeeting` freshness indicator.
function lastMeetingTone(text: string): 'stale' | 'warning' | 'fresh' | null {
  if (!text) return null
  const time = parseISO(text)
  if (!isValid(time)) return null
  const now = new Date()
  let tone: 'stale' | 'warning' | 'fresh' = 'stale'
  if (isAfter(time, subDays(now, STALE_AFTER_DAYS))) tone = 'warning'
  if (isAfter(time, subDays(now, FRESH_WITHIN_DAYS))) tone = 'fresh'
  return tone
}

const toneClasses: Record<'stale' | 'warning' | 'fresh', string> = {
  stale: 'bg-red-100 text-red-700',
  warning: 'bg-amber-100 text-amber-800',
  fresh: 'bg-emerald-100 text-emerald-700',
}

export function CircleTable({
  circles,
  mode,
  isMembersVisible,
  onEdit,
  onDelete,
}: {
  circles: CircleGroup[]
  mode: CircleMode
  isMembersVisible: boolean
  onEdit: (circle: CircleGroup) => void
  onDelete: (circle: CircleGroup) => void
}) {
  const [sorting, setSorting] = useState<SortingState>([
    { id: 'name', desc: false },
  ])

  const columns = useMemo<ColumnDef<CircleGroup>[]>(() => {
    const lastEventColumn: ColumnDef<CircleGroup> = {
      id: 'lastMeeting',
      header: ({ column }) => (
        <button
          type="button"
          onClick={column.getToggleSortingHandler()}
          className="flex items-center gap-1"
        >
          <span>Last Event</span>
          <SortIndicator sorted={column.getIsSorted()} />
        </button>
      ),
      accessorKey: 'last_meeting',
      cell: ({ getValue }) => {
        const lastMeeting = getValue<string>()
        const tone = lastMeetingTone(lastMeeting)
        return (
          <span
            className={cn(
              'rounded-full px-2 py-0.5 text-xs font-medium',
              tone ? toneClasses[tone] : 'bg-muted text-muted-foreground',
            )}
          >
            {lastMeeting || 'None'}
          </span>
        )
      },
    }

    const membersColumn: ColumnDef<CircleGroup> = {
      id: 'members',
      header: 'Members',
      cell: ({ row }) => {
        const members = row.original.members.filter((m) => !m.point_person)
        if (members.length === 0) return null
        return (
          <ul className="list-disc space-y-0.5 pl-4">
            {members.map((m) => (
              <li key={m.name}>{m.name}</li>
            ))}
          </ul>
        )
      },
    }

    const totalMembersColumn: ColumnDef<CircleGroup> = {
      id: 'totalMembers',
      header: 'Total Members',
      accessorFn: (row) => countMailingListMembers(row.members),
    }

    return [
      {
        id: 'actions',
        header: 'Actions',
        cell: ({ row }) => (
          <div className="flex gap-1">
            <Button
              type="button"
              variant="outline"
              size="icon"
              aria-label={`Edit ${row.original.name}`}
              onClick={() => onEdit(row.original)}
            >
              <Pencil className="h-4 w-4 text-primary" />
            </Button>
            <Button
              type="button"
              variant="outline"
              size="icon"
              aria-label={`Delete ${row.original.name}`}
              onClick={() => onDelete(row.original)}
            >
              <Trash2 className="h-4 w-4 text-destructive" />
            </Button>
          </div>
        ),
      },
      {
        id: 'name',
        header: ({ column }) => (
          <button
            type="button"
            onClick={column.getToggleSortingHandler()}
            className="flex items-center gap-1"
          >
            <span>Name</span>
            <SortIndicator sorted={column.getIsSorted()} />
          </button>
        ),
        accessorKey: 'name',
        cell: ({ getValue }) => (
          <span className="font-medium">{getValue<string>()}</span>
        ),
      },
      {
        id: 'host',
        header: 'Host',
        accessorFn: (row) => findPointPerson(row.members)?.name ?? '',
      },
      ...(mode === 'interest'
        ? [lastEventColumn]
        : [isMembersVisible ? membersColumn : totalMembersColumn]),
    ]
  }, [mode, isMembersVisible, onEdit, onDelete])

  // eslint-disable-next-line react-hooks/incompatible-library -- Remove once TanStack Table supports React Compiler.
  const table = useReactTable({
    data: circles,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    onSortingChange: setSorting,
    state: { sorting },
  })

  const rows = table.getRowModel().rows

  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          {table.getHeaderGroups().map((headerGroup) => (
            <TableRow key={headerGroup.id}>
              {headerGroup.headers.map((header) => (
                <TableHead key={header.id} className="whitespace-nowrap">
                  {header.isPlaceholder
                    ? null
                    : flexRender(
                        header.column.columnDef.header,
                        header.getContext(),
                      )}
                </TableHead>
              ))}
            </TableRow>
          ))}
        </TableHeader>
        <TableBody>
          {rows.length ? (
            rows.map((row) => (
              <TableRow key={row.id}>
                {row.getVisibleCells().map((cell) => (
                  <TableCell key={cell.id}>
                    {flexRender(cell.column.columnDef.cell, cell.getContext())}
                  </TableCell>
                ))}
              </TableRow>
            ))
          ) : (
            <TableRow>
              <TableCell
                colSpan={columns.length}
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
