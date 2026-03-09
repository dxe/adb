'use client'

import { Fragment, useMemo, useState } from 'react'
import {
  ColumnDef,
  ExpandedState,
  flexRender,
  getCoreRowModel,
  getExpandedRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable,
} from '@tanstack/react-table'
import {
  ChevronDown,
  ChevronRight,
  Download,
  Mail,
  Pencil,
  Trash2,
} from 'lucide-react'
import { EventListItem } from '@/lib/api'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { IntentPrefetchLink } from '@/components/intent-prefetch-link'
import { EventListMode } from './events-page'

function SortIndicator({ sorted }: { sorted: false | 'asc' | 'desc' }) {
  return (
    <span className={`text-xs${sorted ? '' : ' invisible'}`}>
      {sorted === 'desc' ? '▼' : '▲'}
    </span>
  )
}

export function EventListTable({
  events,
  mode,
  onDelete,
}: {
  events: EventListItem[]
  mode: EventListMode
  onDelete: (event: EventListItem) => void
}) {
  const isConnections = mode === 'connections'
  const [sorting, setSorting] = useState<SortingState>([
    { id: 'event_date', desc: true },
  ])
  const [expanded, setExpanded] = useState<ExpandedState>({})

  const columns = useMemo<ColumnDef<EventListItem>[]>(() => {
    const cols: ColumnDef<EventListItem>[] = [
      {
        id: 'expand',
        header: '',
        cell: ({ row }) => (
          <button
            type="button"
            onClick={row.getToggleExpandedHandler()}
            className="text-muted-foreground hover:text-foreground"
            aria-label={row.getIsExpanded() ? 'Collapse row' : 'Expand row'}
          >
            {row.getIsExpanded() ? (
              <ChevronDown className="h-4 w-4" />
            ) : (
              <ChevronRight className="h-4 w-4" />
            )}
          </button>
        ),
      },
      {
        id: 'actions',
        header: '',
        cell: ({ row }) => (
          <div className="flex items-center gap-1">
            <Button asChild variant="outline" size="sm">
              <IntentPrefetchLink
                href={
                  isConnections
                    ? `/coachings/${row.original.event_id}`
                    : `/events/${row.original.event_id}`
                }
              >
                <Pencil className="h-3.5 w-3.5" />
                Edit
              </IntentPrefetchLink>
            </Button>
            <Button
              variant="outline"
              size="sm"
              className="text-destructive hover:text-destructive"
              aria-label={`Delete event: ${row.original.event_name}`}
              onClick={() => onDelete(row.original)}
            >
              <Trash2 className="h-3.5 w-3.5" />
            </Button>
          </div>
        ),
      },
      {
        id: 'event_date',
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
        accessorKey: 'event_date',
      },
      {
        id: 'event_name',
        header: ({ column }) => (
          <button
            type="button"
            onClick={column.getToggleSortingHandler()}
            className="flex items-center gap-1"
          >
            <span>{isConnections ? 'Coach' : 'Name'}</span>
            <SortIndicator sorted={column.getIsSorted()} />
          </button>
        ),
        accessorKey: 'event_name',
        cell: ({ getValue }) => (
          <span className="font-medium">{getValue<string>()}</span>
        ),
      },
    ]

    if (!isConnections) {
      cols.push(
        {
          id: 'event_type',
          header: ({ column }) => (
            <button
              type="button"
              onClick={column.getToggleSortingHandler()}
              className="flex items-center gap-1"
            >
              <span>Type</span>
              <SortIndicator sorted={column.getIsSorted()} />
            </button>
          ),
          accessorKey: 'event_type',
        },
        {
          id: 'attendee_count',
          header: ({ column }) => (
            <button
              type="button"
              onClick={column.getToggleSortingHandler()}
              className="flex items-center gap-1"
            >
              <span>Total Attendees</span>
              <SortIndicator sorted={column.getIsSorted()} />
            </button>
          ),
          accessorFn: (row) => row.attendees?.length ?? 0,
          cell: ({ getValue }) => getValue<number>(),
        },
      )
    } else {
      cols.push({
        id: 'coachees',
        header: 'Coachees',
        accessorFn: (row) => row.attendees?.join(', ') ?? '',
      })
    }

    return cols
  }, [isConnections, onDelete])

  const table = useReactTable({
    data: events,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getExpandedRowModel: getExpandedRowModel(),
    getRowCanExpand: () => true,
    onSortingChange: setSorting,
    onExpandedChange: setExpanded,
    state: { sorting, expanded },
  })

  const rows = table.getRowModel().rows

  return (
    <div className="space-y-4">
      {/* Desktop table */}
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
                <Fragment key={row.id}>
                  <TableRow>
                    {row.getVisibleCells().map((cell) => (
                      <TableCell key={cell.id}>
                        {flexRender(
                          cell.column.columnDef.cell,
                          cell.getContext(),
                        )}
                      </TableCell>
                    ))}
                  </TableRow>
                  {row.getIsExpanded() && (
                    <TableRow className="bg-muted hover:bg-muted">
                      <TableCell colSpan={columns.length} className="py-4 px-6">
                        <ExpandedDetail
                          event={row.original}
                          isConnections={isConnections}
                        />
                      </TableCell>
                    </TableRow>
                  )}
                </Fragment>
              ))
            ) : (
              <TableRow>
                <TableCell
                  colSpan={columns.length}
                  className="text-center py-6"
                >
                  No {isConnections ? 'coachings' : 'events'} found.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      {/* Mobile cards */}
      <div className="md:hidden space-y-3">
        {rows.length ? (
          rows.map((row) => (
            <div
              key={row.id}
              className="rounded-lg border bg-card p-4 shadow-sm text-card-foreground"
            >
              <div className="flex items-start justify-between gap-3">
                <div className="space-y-0.5">
                  <span className="text-base font-semibold">
                    {row.original.event_name}
                  </span>
                  <span className="block text-sm text-muted-foreground">
                    {row.original.event_date}
                    {!isConnections && ` · ${row.original.event_type}`}
                  </span>
                </div>
                <div className="flex items-center gap-1 shrink-0">
                  <Button asChild variant="outline" size="sm">
                    <IntentPrefetchLink
                      href={
                        isConnections
                          ? `/coachings/${row.original.event_id}`
                          : `/events/${row.original.event_id}`
                      }
                    >
                      <Pencil className="h-3.5 w-3.5" />
                      Edit
                    </IntentPrefetchLink>
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    className="text-destructive hover:text-destructive"
                    aria-label={`Delete event: ${row.original.event_name}`}
                    onClick={() => onDelete(row.original)}
                  >
                    <Trash2 className="h-3.5 w-3.5" />
                  </Button>
                </div>
              </div>
              <dl className="mt-3 text-sm space-y-1">
                {!isConnections && (
                  <div className="flex gap-2">
                    <dt className="text-muted-foreground">Attendees:</dt>
                    <dd>{row.original.attendees?.length ?? 0}</dd>
                  </div>
                )}
                {isConnections && row.original.attendees?.length > 0 && (
                  <div className="flex gap-2">
                    <dt className="text-muted-foreground">Coachees:</dt>
                    <dd>{row.original.attendees.join(', ')}</dd>
                  </div>
                )}
              </dl>
              <div className="mt-3">
                <button
                  type="button"
                  onClick={row.getToggleExpandedHandler()}
                  className="text-sm text-primary flex items-center gap-1"
                >
                  {row.getIsExpanded() ? (
                    <ChevronDown className="h-3.5 w-3.5" />
                  ) : (
                    <ChevronRight className="h-3.5 w-3.5" />
                  )}
                  {row.getIsExpanded() ? 'Hide details' : 'Show details'}
                </button>
                {row.getIsExpanded() && (
                  <div className="mt-3">
                    <ExpandedDetail
                      event={row.original}
                      isConnections={isConnections}
                    />
                  </div>
                )}
              </div>
            </div>
          ))
        ) : (
          <p className="text-center text-sm text-muted-foreground">
            No {isConnections ? 'coachings' : 'events'} found.
          </p>
        )}
      </div>
    </div>
  )
}

function ExpandedDetail({
  event,
  isConnections,
}: {
  event: EventListItem
  isConnections: boolean
}) {
  const emailLink = event.attendee_emails?.length
    ? `https://mail.google.com/mail/?view=cm&fs=1&bcc=${event.attendee_emails.map(encodeURIComponent).join(',')}`
    : null
  const csvLink = `/csv/event_attendance/${event.event_id}`

  const hasAttendees = !isConnections && event.attendees?.length > 0

  return (
    <div className="flex flex-col gap-3 md:flex-row md:items-start md:gap-8">
      {hasAttendees && (
        <div className="flex-1 min-w-0">
          <p className="text-sm font-semibold text-primary mb-1">Attendees</p>
          <ul className="text-sm space-y-0.5 list-disc list-inside text-muted-foreground">
            {event.attendees.map((name) => (
              <li key={name}>{name}</li>
            ))}
          </ul>
        </div>
      )}

      <div className="flex flex-wrap gap-2 md:flex-col md:items-start md:shrink-0">
        {emailLink && (
          <Button asChild variant="outline" size="sm">
            <a href={emailLink} target="_blank" rel="noreferrer">
              <Mail className="h-3.5 w-3.5" />
              Email all attendees
            </a>
          </Button>
        )}
        <Button asChild variant="outline" size="sm">
          <a href={csvLink} target="_blank" rel="noreferrer">
            <Download className="h-3.5 w-3.5" />
            Export attendee CSV
          </a>
        </Button>
      </div>
    </div>
  )
}
