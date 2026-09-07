'use client'

import { useMemo, useState } from 'react'
import {
  ColumnDef,
  columnVisibilityFeature,
  createSortedRowModel,
  flexRender,
  rowSortingFeature,
  SortingState,
  tableFeatures,
  useTable,
} from '@tanstack/react-table'
import { AUTO_SORT_FNS } from '@/lib/table-sort-fns'
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

const features = tableFeatures({
  columnVisibilityFeature,
  rowSortingFeature,
  sortedRowModel: createSortedRowModel(),
  sortFns: AUTO_SORT_FNS,
})

function composeChapterEmail(chapter: ChapterAdmin) {
  const link = buildChapterEmailLink(chapter)
  if (link == null) {
    toast.error(`There are no email addresses listed for ${chapter.Name}!`)
    return
  }
  window.open(link, '_blank', 'noopener,noreferrer')
}

export function ChapterTable({
  chapters,
  showFacebookColumns,
  onDelete,
  isDeleting,
}: {
  chapters: ChapterAdmin[]
  showFacebookColumns: boolean
  onDelete: (chapter: ChapterAdmin) => void
  isDeleting: boolean
}) {
  const [sorting, setSorting] = useState<SortingState>([
    { id: 'Name', desc: false },
  ])

  const columns = useMemo<ColumnDef<typeof features, ChapterAdmin>[]>(() => {
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
              disabled={isDeleting}
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
          ] satisfies ColumnDef<typeof features, ChapterAdmin>[])
        : []),
    ]
  }, [showFacebookColumns, onDelete, isDeleting])

  const table = useTable({
    features,
    data: chapters,
    columns,
    state: { sorting },
    onSortingChange: setSorting,
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
                  No chapters found.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      {/* Mobile cards */}
      <div className="md:hidden space-y-3">
        {rows.length ? (
          rows.map(({ id, original: chapter }) => (
            <div
              key={id}
              className="rounded-lg border bg-card p-4 shadow-sm text-card-foreground"
            >
              <div className="flex items-start justify-between gap-3">
                <span className="text-base font-semibold">
                  {chapter.Flag} {chapter.Name}
                </span>
                <div className="flex items-center gap-1 shrink-0">
                  <Button asChild variant="outline" size="icon">
                    <IntentPrefetchLink
                      href={`/chapters/${chapter.ChapterID}`}
                      aria-label={`Edit ${chapter.Name}`}
                    >
                      <Pencil className="h-4 w-4" />
                    </IntentPrefetchLink>
                  </Button>
                  <Button
                    variant="outline"
                    size="icon"
                    aria-label={`Email ${chapter.Name}`}
                    onClick={() => composeChapterEmail(chapter)}
                  >
                    <Mail className="h-4 w-4" />
                  </Button>
                  <Button
                    variant="outline"
                    size="icon"
                    aria-label={`Delete ${chapter.Name}`}
                    disabled={isDeleting}
                    onClick={() => onDelete(chapter)}
                  >
                    <Trash2 className="h-4 w-4 text-destructive" />
                  </Button>
                </div>
              </div>
              <dl className="mt-3 grid grid-cols-1 gap-2 text-sm sm:grid-cols-2">
                <div className="flex gap-2">
                  <dt className="text-muted-foreground">Mentor:</dt>
                  <dd>{chapter.Mentor || 'None'}</dd>
                </div>
                <div className="flex gap-2">
                  <dt className="text-muted-foreground">Last Contact:</dt>
                  <dd>{chapter.LastContact || 'None'}</dd>
                </div>
                <div className="flex gap-2">
                  <dt className="text-muted-foreground">Last Action:</dt>
                  <dd
                    className={cn(
                      'rounded-full px-2 py-0.5 text-xs',
                      STATUS_COLOR_CLASSES[colorLastAction(chapter.LastAction)],
                    )}
                    title={lastActionTooltip(chapter.LastAction)}
                  >
                    {chapter.LastAction || 'None'}
                  </dd>
                </div>
                {showFacebookColumns && (
                  <>
                    <div className="flex gap-2">
                      <dt className="text-muted-foreground">Last FB Event:</dt>
                      <dd>{chapter.LastFBEvent || 'None'}</dd>
                    </div>
                    <div className="flex items-center gap-2">
                      <dt className="text-muted-foreground">FB Sync Status:</dt>
                      <dd>
                        <span
                          className={cn(
                            'inline-block h-3 w-3 rounded-full',
                            STATUS_COLOR_CLASSES[
                              colorFBSyncStatus(chapter.LastFBSync)
                            ],
                          )}
                        />
                      </dd>
                    </div>
                  </>
                )}
              </dl>
            </div>
          ))
        ) : (
          <p className="text-center text-sm text-muted-foreground">
            No chapters found.
          </p>
        )}
      </div>
    </div>
  )
}
