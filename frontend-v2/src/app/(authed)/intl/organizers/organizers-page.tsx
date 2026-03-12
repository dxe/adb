'use client'

import { useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  ColumnDef,
  VisibilityState,
  flexRender,
  getCoreRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable,
} from '@tanstack/react-table'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Checkbox } from '@/components/ui/checkbox'
import { SortIndicator } from '@/components/ui/sort-indicator'
import {
  apiClient,
  CHAPTER_ORGANIZERS_QUERY_KEY,
  flattenChapterOrganizers,
  type InternationalOrganizer,
} from '@/lib/api'

const SOCIAL_KEYS: (keyof InternationalOrganizer)[] = [
  'facebook',
  'twitter',
  'instagram',
  'linkedin',
]

const COLUMNS: {
  key: keyof InternationalOrganizer
  label: string
  size: number
  social?: boolean
}[] = [
  { key: 'chapterName', label: 'Chapter', size: 200 },
  { key: 'name', label: 'Name', size: 180 },
  { key: 'email', label: 'Email', size: 240 },
  { key: 'phone', label: 'Phone', size: 150 },
  { key: 'facebook', label: 'Facebook', size: 200, social: true },
  { key: 'twitter', label: 'Twitter', size: 180, social: true },
  { key: 'instagram', label: 'Instagram', size: 180, social: true },
  { key: 'linkedin', label: 'LinkedIn', size: 200, social: true },
]

export default function OrganizersPage() {
  const [sorting, setSorting] = useState<SortingState>([])
  const [showSocial, setShowSocial] = useState(false)

  const columnVisibility = useMemo<VisibilityState>(
    () => Object.fromEntries(SOCIAL_KEYS.map((key) => [key, showSocial])),
    [showSocial],
  )

  const {
    data: chapters,
    isLoading,
    isError,
    error,
  } = useQuery({
    queryKey: [...CHAPTER_ORGANIZERS_QUERY_KEY],
    queryFn: ({ signal }) => apiClient.getChapterListWithOrganizers(signal),
  })

  const organizers = useMemo(
    () => (chapters ? flattenChapterOrganizers(chapters) : []),
    [chapters],
  )

  const columns = useMemo<ColumnDef<InternationalOrganizer>[]>(
    () =>
      COLUMNS.map(({ key, label, size }) => ({
        id: key,
        accessorKey: key,
        size,
        minSize: 80,
        header: ({ column }) => (
          <button
            type="button"
            aria-label={label}
            className="flex items-center gap-1 font-medium hover:text-foreground transition-colors truncate"
            onClick={() => column.toggleSorting()}
          >
            {label}
            <SortIndicator
              sorted={
                column.getIsSorted() === 'asc'
                  ? 'asc'
                  : column.getIsSorted() === 'desc'
                    ? 'desc'
                    : false
              }
            />
          </button>
        ),
        cell: ({ getValue }) => (
          <div className="truncate text-sm">{String(getValue() ?? '')}</div>
        ),
      })),
    [],
  )

  const table = useReactTable({
    data: organizers,
    columns,
    state: { sorting, columnVisibility },
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    columnResizeMode: 'onChange',
  })

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-1">
        <h1 className="text-2xl font-semibold">International Organizers</h1>
      </div>

      <label className="flex items-center gap-2 text-sm">
        <Checkbox
          checked={showSocial}
          onCheckedChange={(checked) => setShowSocial(checked === true)}
        />
        Show social media fields
      </label>

      {isLoading && (
        <div className="flex items-center justify-center py-12 text-muted-foreground">
          Loading organizers...
        </div>
      )}

      {isError && (
        <div className="flex items-center justify-center py-12 text-destructive">
          {error instanceof Error
            ? error.message
            : 'Failed to load organizers. Please try again.'}
        </div>
      )}

      {!isLoading && !isError && (
        <>
          {organizers.length > 0 && (
            <div className="text-sm text-muted-foreground">
              {organizers.length} organizer
              {organizers.length !== 1 ? 's' : ''} across{' '}
              {new Set(organizers.map((o) => o.chapterId)).size} chapters
            </div>
          )}

          {organizers.length === 0 ? (
            <div className="flex items-center justify-center py-12 text-muted-foreground">
              No organizers found.
            </div>
          ) : (
            <div className="mx-auto max-w-full overflow-x-auto rounded-md border">
              <Table
                className="table-fixed"
                style={{ width: table.getTotalSize() }}
              >
                <TableHeader>
                  {table.getHeaderGroups().map((headerGroup) => (
                    <TableRow key={headerGroup.id}>
                      {headerGroup.headers.map((header) => (
                        <TableHead
                          key={header.id}
                          className="relative overflow-hidden"
                          style={{ width: header.getSize() }}
                        >
                          {header.isPlaceholder
                            ? null
                            : flexRender(
                                header.column.columnDef.header,
                                header.getContext(),
                              )}
                          <div
                            onMouseDown={header.getResizeHandler()}
                            onTouchStart={header.getResizeHandler()}
                            onDoubleClick={() => header.column.resetSize()}
                            className={`absolute right-0 top-0 h-full w-1 cursor-col-resize select-none touch-none hover:bg-primary/50 ${
                              header.column.getIsResizing() ? 'bg-primary' : ''
                            }`}
                          />
                        </TableHead>
                      ))}
                    </TableRow>
                  ))}
                </TableHeader>
                <TableBody>
                  {table.getRowModel().rows.map((row) => (
                    <TableRow key={row.id}>
                      {row.getVisibleCells().map((cell) => (
                        <TableCell
                          key={cell.id}
                          className="overflow-hidden"
                          style={{ width: cell.column.getSize() }}
                        >
                          {flexRender(
                            cell.column.columnDef.cell,
                            cell.getContext(),
                          )}
                        </TableCell>
                      ))}
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </>
      )}
    </div>
  )
}
