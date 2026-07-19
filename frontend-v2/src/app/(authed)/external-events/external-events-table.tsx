'use client'

import { useMemo, useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable,
} from '@tanstack/react-table'
import { format } from 'date-fns'
import { Star, Trash2 } from 'lucide-react'
import toast from 'react-hot-toast'
import { API_PATH, apiClient, ExternalEvent } from '@/lib/api'
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
import { CancelEventDialog } from './cancel-event-dialog'

export function ExternalEventsTable({ events }: { events: ExternalEvent[] }) {
  const queryClient = useQueryClient()
  const [sorting, setSorting] = useState<SortingState>([
    { id: 'StartTime', desc: false },
  ])
  const [eventToCancel, setEventToCancel] = useState<ExternalEvent | null>(null)

  const featureMutation = useMutation({
    mutationFn: ({ id, featured }: { id: string; featured: boolean }) =>
      apiClient.featureExternalEvent(id, featured),
    onSuccess: (_data, { id, featured }) => {
      toast.success(
        featured
          ? 'Successfully featured event.'
          : 'Successfully unfeatured event.',
      )
      queryClient.setQueryData(
        [API_PATH.EXTERNAL_EVENTS_LIST],
        (old: ExternalEvent[] | undefined) =>
          old?.map((event) =>
            event.ID === id ? { ...event, Featured: featured } : event,
          ),
      )
    },
    onError: (err: Error) => {
      toast.error(err.message || 'Failed to update event')
    },
  })

  const columns = useMemo<ColumnDef<ExternalEvent>[]>(
    () => [
      {
        id: 'StartTime',
        header: ({ column }) => (
          <button
            type="button"
            onClick={column.getToggleSortingHandler()}
            className="flex items-center gap-1"
          >
            <span>Date</span>
            <SortIndicator sorted={column.getIsSorted()} />
          </button>
        ),
        accessorKey: 'StartTime',
        cell: ({ getValue }) =>
          format(new Date(getValue<string>()), 'yyyy-MM-dd'),
      },
      {
        id: 'Name',
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
        accessorKey: 'Name',
        cell: ({ row }) => (
          <a
            href={`https://facebook.com/${row.original.ID}`}
            target="_blank"
            rel="noreferrer"
            className="text-primary underline-offset-4 hover:underline"
          >
            {row.original.Name}
          </a>
        ),
      },
      {
        id: 'featured',
        header: '',
        cell: ({ row }) => {
          const event = row.original
          const isPending =
            featureMutation.isPending &&
            featureMutation.variables?.id === event.ID
          return (
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() =>
                featureMutation.mutate({
                  id: event.ID,
                  featured: !event.Featured,
                })
              }
              disabled={isPending}
              className={
                event.Featured
                  ? 'text-amber-600 hover:text-amber-600'
                  : undefined
              }
            >
              <Star className={event.Featured ? 'fill-current' : undefined} />
              {event.Featured ? 'Featured' : 'Feature'}
            </Button>
          )
        },
      },
      {
        id: 'actions',
        header: '',
        cell: ({ row }) => (
          <Button
            type="button"
            variant="destructive"
            size="sm"
            onClick={() => setEventToCancel(row.original)}
          >
            <Trash2 />
            Delete
          </Button>
        ),
      },
    ],
    [featureMutation],
  )

  // eslint-disable-next-line react-hooks/incompatible-library -- Remove once TanStack Table supports React Compiler.
  const table = useReactTable({
    data: events,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    onSortingChange: setSorting,
    state: {
      sorting,
    },
  })

  const rows = table.getRowModel().rows

  return (
    <div className="space-y-4">
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
                  No upcoming events found.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      {eventToCancel && (
        <CancelEventDialog
          open
          onOpenChange={(open) => {
            if (!open) setEventToCancel(null)
          }}
          eventId={eventToCancel.ID}
          eventName={eventToCancel.Name}
        />
      )}
    </div>
  )
}
