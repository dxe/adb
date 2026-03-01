'use client'

import { useMemo, useState } from 'react'
import Link from 'next/link'
import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable,
} from '@tanstack/react-table'
import { User } from '@/lib/api'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import clsx from 'clsx'

type Chapter = {
  ChapterID: number
  Name: string
}

export function UserTable({
  users,
  chapters,
}: {
  users: User[]
  chapters: Chapter[]
}) {
  const [sorting, setSorting] = useState<SortingState>([])

  const chapterMap = useMemo(
    () => new Map(chapters.map((c) => [c.ChapterID, c.Name])),
    [chapters],
  )

  const columns = useMemo<ColumnDef<User>[]>(() => {
    const sortIndicator = (state: false | 'asc' | 'desc') => {
      if (state === 'asc') return '▲'
      if (state === 'desc') return '▼'
      return ''
    }

    return [
      {
        header: ({ column }) => (
          <button
            type="button"
            onClick={column.getToggleSortingHandler()}
            className="flex items-center gap-1"
          >
            <span>Name</span>
            <span className="text-xs">
              {sortIndicator(column.getIsSorted())}
            </span>
          </button>
        ),
        accessorKey: 'name',
        cell: ({ row }) => (
          <Link
            href={`/users/${row.original.id}`}
            prefetch={false}
            className="font-semibold hover:underline"
          >
            {row.original.name}
          </Link>
        ),
      },
      {
        header: ({ column }) => (
          <button
            type="button"
            onClick={column.getToggleSortingHandler()}
            className="flex items-center gap-1"
          >
            <span>Email</span>
            <span className="text-xs">
              {sortIndicator(column.getIsSorted())}
            </span>
          </button>
        ),
        accessorKey: 'email',
        cell: ({ getValue }) => (
          <span className="font-mono text-sm">{getValue<string>()}</span>
        ),
      },
      {
        id: 'chapter',
        header: ({ column }) => (
          <button
            type="button"
            onClick={column.getToggleSortingHandler()}
            className="flex items-center gap-1"
          >
            <span>Chapter</span>
            <span className="text-xs">
              {sortIndicator(column.getIsSorted())}
            </span>
          </button>
        ),
        accessorFn: (row) => chapterMap.get(row.chapter_id) ?? '',
        cell: ({ getValue }) => {
          const chapterName = getValue<string>()
          return chapterName || 'Unknown'
        },
      },
      {
        header: 'Roles',
        accessorKey: 'roles',
        cell: ({ getValue }) => {
          const roles = getValue<string[]>()
          if (!roles.length) return '—'
          return roles.join(', ')
        },
      },
      {
        id: 'status',
        header: ({ column }) => (
          <button
            type="button"
            onClick={column.getToggleSortingHandler()}
            className="flex items-center gap-1"
          >
            <span>Status</span>
            <span className="text-xs">
              {sortIndicator(column.getIsSorted())}
            </span>
          </button>
        ),
        accessorFn: (row) => (row.disabled ? 'Disabled' : 'Active'),
        cell: ({ getValue, row }) => {
          const status = getValue<string>()
          const isDisabled = row.original.disabled
          return (
            <span
              className={clsx(
                'text-xs font-semibold',
                isDisabled ? 'text-destructive' : 'text-emerald-700',
              )}
            >
              {status}
            </span>
          )
        },
      },
    ]
  }, [chapterMap])

  const table = useReactTable({
    data: users,
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
                  No users found.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      <div className="md:hidden space-y-3">
        {rows.length ? (
          rows.map((row) => (
            <div
              key={row.id}
              className="rounded-lg border bg-card p-4 shadow-sm text-card-foreground"
            >
              <div className="flex items-start justify-between gap-3">
                <div className="space-y-1">
                  <Link
                    href={`/users/${row.original.id}`}
                    prefetch={false}
                    className="text-base font-semibold hover:underline"
                  >
                    {row.original.name}
                  </Link>
                  <span className="block text-sm text-muted-foreground">
                    {row.original.email}
                  </span>
                </div>
              </div>
              <dl className="mt-3 grid grid-cols-1 gap-2 text-sm sm:grid-cols-2">
                <div className="flex gap-2">
                  <dt className="text-muted-foreground">Chapter:</dt>
                  <dd>
                    {chapterMap.get(row.original.chapter_id) ?? 'Unknown'}
                  </dd>
                </div>
                <div className="flex gap-2">
                  <dt className="text-muted-foreground">Status:</dt>
                  <dd
                    className={
                      row.original.disabled
                        ? 'font-semibold text-destructive'
                        : 'font-semibold text-emerald-700'
                    }
                  >
                    {row.original.disabled ? 'Disabled' : 'Active'}
                  </dd>
                </div>
                <div className="flex gap-2 sm:col-span-2">
                  <dt className="text-muted-foreground">Roles:</dt>
                  <dd>
                    {row.original.roles.length
                      ? row.original.roles.join(', ')
                      : '—'}
                  </dd>
                </div>
              </dl>
            </div>
          ))
        ) : (
          <p className="text-center text-sm text-muted-foreground">
            No users found.
          </p>
        )}
      </div>
    </div>
  )
}
