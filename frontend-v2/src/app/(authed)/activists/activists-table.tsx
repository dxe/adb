'use client'

import { useMemo } from 'react'
import {
  ColumnDef,
  columnResizingFeature,
  columnSizingFeature,
  columnVisibilityFeature,
  flexRender,
  tableFeatures,
  useTable,
} from '@tanstack/react-table'
import {
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { StickyHeaderTable } from '@/components/sticky-header-table'
import { ArrowDown, ArrowUp, Check, Minus } from 'lucide-react'
import { ActivistJSON, ActivistColumnName } from '@/lib/api'
import { Checkbox } from '@/components/ui/checkbox'
import { ActivistCard } from './activist-card'
import { COLUMN_DEFINITION_BY_NAME } from './column-definitions'
import { getActivistDisplayName } from './display-name'
import { formatValue, COLUMN_TYPE_BY_NAME } from './format-value'
import type { SortColumn } from './query-state'
import type { ActivistSelection } from './use-activist-selection'

const features = tableFeatures({
  columnSizingFeature,
  columnResizingFeature,
  columnVisibilityFeature,
})

// Id of the leading checkbox column, which holds no activist data and is not
// resizable.
const SELECT_COLUMN_ID = '__select'

/**
 * The leading checkbox column. Its header checkbox covers the loaded rows only,
 * which is all "select all" can reach — later pages are not fetched yet.
 */
function buildSelectColumn(
  activists: ActivistJSON[],
  selection: ActivistSelection,
): ColumnDef<typeof features, ActivistJSON> {
  const rowIds = activists.map((activist) => activist.id)
  const selectedCount = rowIds.filter((id) =>
    selection.selectedIds.has(id),
  ).length
  const headerState =
    selectedCount === 0
      ? false
      : selectedCount === rowIds.length
        ? true
        : 'indeterminate'

  return {
    id: SELECT_COLUMN_ID,
    size: 44,
    minSize: 44,
    header: () => (
      <Checkbox
        checked={headerState}
        onCheckedChange={(checked) =>
          selection.onSetMany(rowIds, checked === true)
        }
        aria-label={
          headerState === true
            ? 'Deselect all activists'
            : 'Select all activists'
        }
      />
    ),
    cell: ({ row }) => (
      <Checkbox
        checked={selection.selectedIds.has(row.original.id)}
        onCheckedChange={() => selection.onToggle(row.original.id)}
        aria-label={`Select ${getActivistDisplayName(row.original).text}`}
      />
    ),
  }
}

interface ActivistTableProps {
  activists: ActivistJSON[]
  visibleColumns: ActivistColumnName[]
  sort: SortColumn[]
  onSortChange: (sort: SortColumn[]) => void
  onActivistClick?: (id: number) => void
  isStale?: boolean
  footer?: React.ReactNode
  /** Omit to render the table without any selection affordances. */
  selection?: ActivistSelection
}

export function ActivistTable({
  activists,
  visibleColumns,
  sort,
  onSortChange,
  onActivistClick,
  isStale = false,
  footer,
  selection,
}: ActivistTableProps) {
  const columns = useMemo<ColumnDef<typeof features, ActivistJSON>[]>(() => {
    const dataColumns = visibleColumns.map(
      (colName): ColumnDef<typeof features, ActivistJSON> => {
        const definition = COLUMN_DEFINITION_BY_NAME[colName]
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
            if (colName === 'name') {
              const displayName = getActivistDisplayName(row.original)
              const nameClass = `truncate text-sm text-primary hover:underline ${
                displayName.isPlaceholder ? 'italic text-muted-foreground' : ''
              }`
              return (
                <a
                  href={`/v2/activists/${row.original.id}`}
                  className={nameClass}
                  onClick={
                    onActivistClick
                      ? (e) => {
                          if (e.ctrlKey || e.metaKey || e.shiftKey) return
                          e.preventDefault()
                          onActivistClick(row.original.id)
                        }
                      : undefined
                  }
                >
                  {displayName.text}
                </a>
              )
            }

            const value = row.original[colName as keyof ActivistJSON]
            if (COLUMN_TYPE_BY_NAME[colName] === 'boolean') {
              return (
                <div className="flex items-center text-sm">
                  {value ? (
                    <Check className="h-4 w-4 text-foreground" />
                  ) : (
                    <Minus className="h-4 w-4 text-muted-foreground" />
                  )}
                </div>
              )
            }
            const formatted = formatValue(value, colName)
            return <div className="truncate text-sm">{formatted}</div>
          },
        }
      },
    )

    if (!selection) return dataColumns

    return [buildSelectColumn(activists, selection), ...dataColumns]
  }, [
    visibleColumns,
    sort,
    onSortChange,
    onActivistClick,
    isStale,
    selection,
    activists,
  ])

  const table = useTable({
    features,
    data: activists,
    columns,
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
      {/*
        Bounded-height flex chain link (md+) — see frontend-v2/docs/patterns/bounded-height-flex-chain.md
        `self-start` keeps the table at its natural width (set inline from
        `table.getTotalSize()`) instead of stretching to fill the column.
      */}
      <div
        className={`hidden max-w-full flex-1 min-h-0 flex-col self-start transition-opacity md:flex ${
          isStale ? 'opacity-60' : ''
        }`}
      >
        <StickyHeaderTable
          data-testid="activists-table"
          className="table-fixed"
          style={{ width: table.getTotalSize() }}
          footer={footer}
        >
          <TableHeader className="sticky top-0 z-10 bg-background">
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
                    {header.column.id !== SELECT_COLUMN_ID && (
                      <div
                        onMouseDown={header.getResizeHandler()}
                        onTouchStart={header.getResizeHandler()}
                        onDoubleClick={() => header.column.resetSize()}
                        className={`absolute right-0 top-0 h-full w-1 cursor-col-resize select-none touch-none hover:bg-primary/50 ${
                          header.column.getIsResizing() ? 'bg-primary' : ''
                        }`}
                      />
                    )}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {table.getRowModel().rows.map((row) => {
              const isSelected =
                selection?.selectedIds.has(row.original.id) ?? false
              return (
                <TableRow
                  key={row.id}
                  // `cn` merges away the base row's conflicting hover
                  // background so the highlight survives hovering.
                  className={
                    isSelected ? 'bg-primary/20 hover:bg-primary/25' : undefined
                  }
                >
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
              )
            })}
          </TableBody>
        </StickyHeaderTable>
      </div>

      {/* Mobile card layout */}
      <div className="flex flex-col gap-4 md:hidden">
        {activists.map((activist) => (
          <ActivistCard
            key={activist.id}
            activist={activist}
            visibleColumns={visibleColumns}
            onActivistClick={onActivistClick}
            isStale={isStale}
            selection={selection}
          />
        ))}
      </div>
      {footer && <div className="md:hidden">{footer}</div>}
    </>
  )
}
