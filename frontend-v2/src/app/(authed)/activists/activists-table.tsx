'use client'

import { useMemo } from 'react'
import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  useReactTable,
  ColumnResizeMode,
} from '@tanstack/react-table'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { ArrowDown, ArrowUp } from 'lucide-react'
import { ActivistJSON, ActivistColumnName } from '@/lib/api'
import { COLUMN_DEFINITIONS } from './column-definitions'
import type { SortColumn } from './query-state'
import { z } from 'zod'

interface ActivistTableProps {
  activists: ActivistJSON[]
  visibleColumns: ActivistColumnName[]
  sort: SortColumn[]
  onSortChange: (sort: SortColumn[]) => void
  isStale?: boolean
}

// Gets the underlying type of a column from the ActivistJSON schema
const getColumnType = (
  columnName: ActivistColumnName,
): 'string' | 'number' | 'boolean' => {
  const schema = ActivistJSON.shape[
    columnName as keyof typeof ActivistJSON.shape
  ] as z.ZodTypeAny
  if (!schema) throw new Error('column not in schema: ' + columnName)

  const unwrapped =
    schema instanceof z.ZodOptional || schema instanceof z.ZodNullable
      ? schema.unwrap()
      : schema

  if (unwrapped instanceof z.ZodNumber) return 'number'
  if (unwrapped instanceof z.ZodBoolean) return 'boolean'
  return 'string'
}

const formatValue = (
  value: unknown,
  columnName: ActivistColumnName,
): string => {
  if (value === null || value === undefined) return ''

  const columnType = getColumnType(columnName)

  if (columnType === 'boolean') {
    return value ? 'Yes' : 'No'
  }

  if (columnType === 'number') {
    return String(value)
  }

  if (columnType === 'string') {
    const definition = COLUMN_DEFINITIONS.find((d) => d.name === columnName)
    if (definition?.isDate && typeof value === 'string') {
      const date = new Date(value)
      if (!isNaN(date.getTime())) {
        return new Intl.DateTimeFormat('en-US', {
          year: 'numeric',
          month: 'short',
          day: 'numeric',
          timeZone: 'UTC',
        }).format(date)
      }
    }
  }

  return String(value)
}

export function ActivistTable({
  activists,
  visibleColumns,
  sort,
  onSortChange,
  isStale = false,
}: ActivistTableProps) {
  const columns = useMemo<ColumnDef<ActivistJSON>[]>(() => {
    return visibleColumns.map((colName) => {
      const definition = COLUMN_DEFINITIONS.find((d) => d.name === colName)
      const label = definition?.label || colName
      const sortIndex = sort.findIndex((s) => s.column === colName)
      const sortEntry = sortIndex !== -1 ? sort[sortIndex] : undefined
      const SortIcon = sortEntry?.desc ? ArrowDown : ArrowUp

      const handleHeaderClick = () => {
        if (isStale) return
        if (sort.length === 1 && sort[0].column === colName) {
          // Toggle direction on the sole sort column (id is always ASC for cursor pagination)
          if (colName === 'id') return
          onSortChange([{ column: colName, desc: !sort[0].desc }])
        } else {
          // Replace all sorting with this column ascending
          onSortChange([{ column: colName, desc: false }])
        }
      }

      return {
        id: colName,
        size: definition?.defaultWidth ?? 150,
        minSize: definition?.minWidth ?? 60,
        header: () => (
          <button
            type="button"
            className="flex items-center gap-1 font-medium hover:text-foreground transition-colors truncate"
            onClick={handleHeaderClick}
            disabled={isStale}
          >
            {label}
            {sortEntry && (
              <>
                <SortIcon className="h-3 w-3 shrink-0 text-muted-foreground" />
                {sort.length > 1 && (
                  <span className="flex h-4 min-w-4 shrink-0 items-center justify-center rounded-full bg-muted px-1 text-[10px] font-semibold text-muted-foreground">
                    {sortIndex + 1}
                  </span>
                )}
              </>
            )}
          </button>
        ),
        accessorFn: (row) => row[colName as keyof ActivistJSON],
        cell: ({ row }) => {
          const value = row.original[colName as keyof ActivistJSON]
          return (
            <div className="truncate text-sm">
              {formatValue(value, colName)}
            </div>
          )
        },
      }
    })
  }, [visibleColumns, sort, onSortChange, isStale])

  const table = useReactTable({
    data: activists,
    columns,
    getCoreRowModel: getCoreRowModel(),
    columnResizeMode: 'onChange',
  })

  if (activists.length === 0) {
    return (
      <div className="flex items-center justify-center py-12 text-muted-foreground">
        No activists found matching the current filters.
      </div>
    )
  }

  return (
    <>
      {/* Desktop table */}
      <div
        className={`mx-auto hidden max-w-full overflow-x-auto rounded-md border transition-opacity md:block ${
          isStale ? 'opacity-60' : ''
        }`}
      >
        <Table className="table-fixed" style={{ width: table.getTotalSize() }}>
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
                    {flexRender(cell.column.columnDef.cell, cell.getContext())}
                  </TableCell>
                ))}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>

      {/* Mobile card layout */}
      <div className="flex flex-col gap-4 md:hidden">
        {activists.map((activist) => (
          <div
            key={activist.id}
            className={`rounded-lg border bg-card p-4 transition-opacity ${
              isStale ? 'opacity-60' : ''
            }`}
          >
            <div className="flex flex-col gap-2">
              {visibleColumns.map((colName) => {
                const definition = COLUMN_DEFINITIONS.find(
                  (d) => d.name === colName,
                )
                const label = definition?.label || colName
                const value = activist[colName as keyof ActivistJSON]

                return (
                  <div key={colName} className="flex justify-between gap-2">
                    <span className="text-sm font-medium text-muted-foreground">
                      {label}:
                    </span>
                    <span className="text-sm">
                      {formatValue(value, colName)}
                    </span>
                  </div>
                )
              })}
            </div>
          </div>
        ))}
      </div>
    </>
  )
}
