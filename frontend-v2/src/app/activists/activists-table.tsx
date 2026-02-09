'use client'

import { useMemo } from 'react'
import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
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
import { ActivistJSON, ActivistColumnName } from '@/lib/api'
import { COLUMN_DEFINITIONS } from './column-definitions'

interface ActivistTableProps {
  activists: ActivistJSON[]
  visibleColumns: ActivistColumnName[]
}

// Gets the underlying type of a column from the ActivistJSON schema
const getColumnType = (
  columnName: ActivistColumnName,
): 'string' | 'number' | 'boolean' => {
  const schema =
    ActivistJSON.shape[columnName as keyof typeof ActivistJSON.shape]
  if (!schema) throw new Error('column not in schema: ' + columnName)

  // Unwrap optional to get the base type
  const unwrapped = schema._def.innerType || schema

  // Check the Zod type
  const typeName = unwrapped._def.typeName
  if (typeName === 'ZodNumber') return 'number'
  if (typeName === 'ZodBoolean') return 'boolean'
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
}: ActivistTableProps) {
  const columns = useMemo<ColumnDef<ActivistJSON>[]>(() => {
    return visibleColumns.map((colName) => {
      const definition = COLUMN_DEFINITIONS.find((d) => d.name === colName)
      const label = definition?.label || colName

      return {
        id: colName,
        header: () => <span className="font-medium">{label}</span>,
        accessorFn: (row) => row[colName as keyof ActivistJSON],
        cell: ({ row }) => {
          const value = row.original[colName as keyof ActivistJSON]
          return <div className="text-sm">{formatValue(value, colName)}</div>
        },
      }
    })
  }, [visibleColumns])

  const table = useReactTable({
    data: activists,
    columns,
    getCoreRowModel: getCoreRowModel(),
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
      <div className="hidden overflow-auto rounded-md border md:block">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <TableHead key={header.id}>
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
            {table.getRowModel().rows.map((row) => (
              <TableRow key={row.id}>
                {row.getVisibleCells().map((cell) => (
                  <TableCell key={cell.id}>
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
          <div key={activist.id} className="rounded-lg border bg-card p-4">
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
