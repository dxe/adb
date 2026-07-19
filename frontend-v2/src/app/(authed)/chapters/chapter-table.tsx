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
import toast from 'react-hot-toast'
import { Pencil, Mail, Trash2 } from 'lucide-react'
import { ChapterAdmin } from '@/lib/api'
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
import { IntentPrefetchLink } from '@/components/intent-prefetch-link'
import { cn } from '@/lib/utils'
import {
  buildChapterEmailLink,
  colorFBSyncStatus,
  colorLastAction,
  lastActionTooltip,
  STATUS_COLOR_CLASSES,
} from './chapter-utils'

function composeChapterEmail(chapter: ChapterAdmin) {
  const link = buildChapterEmailLink(chapter)
  if (link == null) {
    toast.error(`There are no email addresses listed for ${chapter.Name}!`)
    return
  }
  window.open(link)
}

export function ChapterTable({
  chapters,
  showFacebookColumns,
  onDelete,
}: {
  chapters: ChapterAdmin[]
  showFacebookColumns: boolean
  onDelete: (chapter: ChapterAdmin) => void
}) {
  const [sorting, setSorting] = useState<SortingState>([
    { id: 'Name', desc: false },
  ])

  const columns = useMemo<ColumnDef<ChapterAdmin>[]>(() => {
    const sortableHeader = (label: string) =>
      function Header({
        column,
      }: {
        column: {
          toggleSorting: () => void
          getIsSorted: () => false | 'asc' | 'desc'
        }
      }) {
        return (
          <button
            type="button"
            className="flex items-center gap-1 font-medium hover:text-foreground transition-colors"
            onClick={() => column.toggleSorting()}
          >
            {label}
            <SortIndicator sorted={column.getIsSorted()} />
          </button>
        )
      }

    return [
      {
        id: 'actions',
        header: '',
        cell: ({ row }) => (
          <div className="flex items-center gap-1">
            <Button asChild variant="outline" size="icon">
              <IntentPrefetchLink
                href={`/chapters/${row.original.ChapterID}`}
                aria-label={`Edit ${row.original.Name}`}
              >
                <Pencil className="h-4 w-4" />
              </IntentPrefetchLink>
            </Button>
            <Button
              variant="outline"
              size="icon"
              aria-label={`Email ${row.original.Name}`}
              onClick={() => composeChapterEmail(row.original)}
            >
              <Mail className="h-4 w-4" />
            </Button>
            <Button
              variant="outline"
              size="icon"
              aria-label={`Delete ${row.original.Name}`}
              onClick={() => onDelete(row.original)}
            >
              <Trash2 className="h-4 w-4 text-destructive" />
            </Button>
          </div>
        ),
      },
      {
        id: 'Name',
        accessorKey: 'Name',
        header: sortableHeader('Name'),
        cell: ({ row }) => (
          <span>
            {row.original.Flag} {row.original.Name}
          </span>
        ),
      },
      {
        id: 'Mentor',
        accessorKey: 'Mentor',
        header: sortableHeader('Mentor'),
      },
      {
        id: 'LastContact',
        accessorKey: 'LastContact',
        header: sortableHeader('Last Contact'),
        cell: ({ getValue }) => (
          <span className="rounded-full bg-gray-100 px-2 py-0.5 text-xs">
            {getValue<string>() || 'None'}
          </span>
        ),
      },
      {
        id: 'LastAction',
        accessorKey: 'LastAction',
        header: sortableHeader('Last Action'),
        cell: ({ getValue }) => {
          const value = getValue<string>()
          return (
            <span
              className={cn(
                'rounded-full px-2 py-0.5 text-xs',
                STATUS_COLOR_CLASSES[colorLastAction(value)],
              )}
              title={lastActionTooltip(value)}
            >
              {value || 'None'}
            </span>
          )
        },
      },
      ...(showFacebookColumns
        ? ([
            {
              id: 'LastFBEvent',
              accessorKey: 'LastFBEvent',
              header: sortableHeader('Last FB Event'),
              cell: ({ getValue }) => (
                <span className="rounded-full bg-gray-100 px-2 py-0.5 text-xs">
                  {getValue<string>() || 'None'}
                </span>
              ),
            },
            {
              id: 'LastFBSync',
              accessorKey: 'LastFBSync',
              header: sortableHeader('FB Sync Status'),
              cell: ({ getValue }) => (
                <span
                  className={cn(
                    'inline-block h-3 w-3 rounded-full',
                    STATUS_COLOR_CLASSES[colorFBSyncStatus(getValue<string>())],
                  )}
                />
              ),
            },
          ] satisfies ColumnDef<ChapterAdmin>[])
        : []),
    ]
  }, [showFacebookColumns, onDelete])

  // eslint-disable-next-line react-hooks/incompatible-library -- Remove once TanStack Table supports React Compiler.
  const table = useReactTable({
    data: chapters,
    columns,
    state: { sorting },
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
  })

  const rows = table.getRowModel().rows

  return (
    <div className="rounded-md border overflow-x-auto">
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
              <TableCell colSpan={columns.length} className="text-center py-6">
                No chapters found.
              </TableCell>
            </TableRow>
          )}
        </TableBody>
      </Table>
    </div>
  )
}
