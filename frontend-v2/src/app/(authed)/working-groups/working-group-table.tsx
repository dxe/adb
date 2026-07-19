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
import { Pencil, Trash2 } from 'lucide-react'
import { WorkingGroup } from '@/lib/api'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { SortIndicator } from '@/components/ui/sort-indicator'
import { countMailingListMembers, findPointPerson } from '@/lib/members'

export function WorkingGroupTable({
  workingGroups,
  membersVisible,
  onEdit,
  onDelete,
}: {
  workingGroups: WorkingGroup[]
  membersVisible: boolean
  onEdit: (workingGroup: WorkingGroup) => void
  onDelete: (workingGroup: WorkingGroup) => void
}) {
  const [sorting, setSorting] = useState<SortingState>([
    { id: 'name', desc: false },
  ])

  const columns = useMemo<ColumnDef<WorkingGroup>[]>(() => {
    const cols: ColumnDef<WorkingGroup>[] = [
      {
        id: 'actions',
        header: '',
        cell: ({ row }) => (
          <div className="flex items-center gap-1">
            <Button
              variant="outline"
              size="sm"
              aria-label={`Edit working group: ${row.original.name}`}
              onClick={() => onEdit(row.original)}
            >
              <Pencil className="h-3.5 w-3.5" />
              Edit
            </Button>
            <Button
              variant="outline"
              size="sm"
              className="text-destructive hover:text-destructive"
              aria-label={`Delete working group: ${row.original.name}`}
              onClick={() => onDelete(row.original)}
            >
              <Trash2 className="h-3.5 w-3.5" />
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
          <span className="font-semibold">{getValue<string>()}</span>
        ),
      },
      {
        id: 'email',
        header: ({ column }) => (
          <button
            type="button"
            onClick={column.getToggleSortingHandler()}
            className="flex items-center gap-1"
          >
            <span>Email</span>
            <SortIndicator sorted={column.getIsSorted()} />
          </button>
        ),
        accessorKey: 'email',
        cell: ({ getValue }) => (
          <span className="font-mono text-sm">{getValue<string>()}</span>
        ),
      },
      {
        id: 'pointPerson',
        header: 'Point Person',
        accessorFn: (row) => findPointPerson(row.members)?.name ?? '',
      },
    ]

    if (membersVisible) {
      cols.push({
        id: 'members',
        header: 'Members',
        cell: ({ row }) => {
          const members = row.original.members.filter((m) => !m.point_person)
          if (members.length === 0) return null
          return (
            <ul className="list-disc list-inside space-y-0.5">
              {members.map((m) => (
                <li key={m.name}>{m.name}</li>
              ))}
            </ul>
          )
        },
      })
    } else {
      cols.push({
        id: 'totalMembers',
        header: 'Total Members',
        accessorFn: (row) => countMailingListMembers(row.members),
      })
    }

    return cols
  }, [membersVisible, onEdit, onDelete])

  // eslint-disable-next-line react-hooks/incompatible-library -- Remove once TanStack Table supports React Compiler.
  const table = useReactTable({
    data: workingGroups,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    onSortingChange: setSorting,
    state: { sorting },
  })

  const rows = table.getRowModel().rows

  return (
    <div className="space-y-4">
      <div className="rounded-md border hidden md:block">
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
                      {flexRender(
                        cell.column.columnDef.cell,
                        cell.getContext(),
                      )}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell
                  colSpan={columns.length}
                  className="text-center py-6"
                >
                  No working groups found.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      <div className="md:hidden space-y-3">
        {rows.length ? (
          rows.map((row) => {
            const workingGroup = row.original
            const members = workingGroup.members.filter((m) => !m.point_person)
            return (
              <div
                key={row.id}
                className="rounded-lg border bg-card p-4 shadow-sm text-card-foreground"
              >
                <div className="flex items-start justify-between gap-3">
                  <div className="space-y-1">
                    <span className="text-base font-semibold">
                      {workingGroup.name}
                    </span>
                    <span className="block text-sm text-muted-foreground font-mono">
                      {workingGroup.email}
                    </span>
                  </div>
                  <div className="flex items-center gap-1 shrink-0">
                    <Button
                      variant="outline"
                      size="sm"
                      aria-label={`Edit working group: ${workingGroup.name}`}
                      onClick={() => onEdit(workingGroup)}
                    >
                      <Pencil className="h-3.5 w-3.5" />
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      className="text-destructive hover:text-destructive"
                      aria-label={`Delete working group: ${workingGroup.name}`}
                      onClick={() => onDelete(workingGroup)}
                    >
                      <Trash2 className="h-3.5 w-3.5" />
                    </Button>
                  </div>
                </div>
                <dl className="mt-3 text-sm space-y-1">
                  <div className="flex gap-2">
                    <dt className="text-muted-foreground">Point Person:</dt>
                    <dd>
                      {findPointPerson(workingGroup.members)?.name || '—'}
                    </dd>
                  </div>
                  {membersVisible ? (
                    members.length > 0 && (
                      <div className="flex gap-2">
                        <dt className="text-muted-foreground">Members:</dt>
                        <dd>{members.map((m) => m.name).join(', ')}</dd>
                      </div>
                    )
                  ) : (
                    <div className="flex gap-2">
                      <dt className="text-muted-foreground">Total Members:</dt>
                      <dd>{countMailingListMembers(workingGroup.members)}</dd>
                    </div>
                  )}
                </dl>
              </div>
            )
          })
        ) : (
          <p className="text-center text-sm text-muted-foreground">
            No working groups found.
          </p>
        )}
      </div>
    </div>
  )
}
