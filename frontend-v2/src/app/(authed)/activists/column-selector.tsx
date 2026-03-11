'use client'

import { useState, useMemo } from 'react'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { CircleHelp, Columns3, Search } from 'lucide-react'
import { ActivistColumnName } from '@/lib/api'
import {
  groupColumnsByCategory,
  ColumnCategory,
  ColumnDefinition,
  normalizeColumns,
} from './column-definitions'

interface ColumnSelectorProps {
  visibleColumns: ActivistColumnName[]
  onColumnsChange: (columns: ActivistColumnName[]) => void
  // True to show "Chapter" as a column that is checked and disabled (cannot be unchecked), and
  // false to hide the "Chapter" column from the list.
  isChapterColumnShown: boolean
}

export function ColumnSelector({
  visibleColumns,
  onColumnsChange,
  isChapterColumnShown,
}: ColumnSelectorProps) {
  const [isOpen, setIsOpen] = useState(false)
  const [localColumns, setLocalColumns] = useState(visibleColumns)
  const [search, setSearch] = useState('')
  const groupedColumns = useMemo(() => groupColumnsByCategory(), [])

  const slugifyCategory = (category: ColumnCategory) =>
    category
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, '-')
      .replace(/^-+|-+$/g, '')

  // Filter categories and columns based on search query
  const isSearching = search.trim().length > 0

  // Filter categories and columns based on search query
  const filteredGroups = useMemo(() => {
    const query = search.trim().toLowerCase()
    if (!query) return groupedColumns

    const filtered = new Map<ColumnCategory, ColumnDefinition[]>()
    for (const [category, columns] of groupedColumns) {
      const categoryMatches = category.toLowerCase().includes(query)
      const matchingColumns = columns.filter(
        (col) =>
          categoryMatches ||
          col.label.toLowerCase().includes(query) ||
          col.name.toLowerCase().includes(query),
      )
      if (matchingColumns.length > 0) {
        filtered.set(category, matchingColumns)
      }
    }
    return filtered
  }, [search, groupedColumns])

  const handleToggleColumn = (columnName: ActivistColumnName) => {
    const newColumns = localColumns.includes(columnName)
      ? localColumns.filter((col) => col !== columnName)
      : [...localColumns, columnName]
    setLocalColumns(normalizeColumns(newColumns))
  }

  const handleToggleCategory = (category: ColumnCategory) => {
    const categoryColumns = groupedColumns.get(category) || []
    const categoryColumnNames = categoryColumns
      .filter((col) => !col.hidden)
      .map((col) => col.name)
      .filter((name) => name !== 'chapter_name') // chapter_name is managed outside of this component
    const allVisible = categoryColumnNames.every((col) =>
      localColumns.includes(col),
    )

    let newColumns: ActivistColumnName[]
    if (allVisible) {
      // Remove all columns in this category
      newColumns = localColumns.filter(
        (col) => !categoryColumnNames.includes(col),
      )
    } else {
      // Add all columns in this category
      newColumns = [...localColumns]
      categoryColumnNames.forEach((col) => {
        if (!newColumns.includes(col)) {
          newColumns.push(col)
        }
      })
    }
    setLocalColumns(normalizeColumns(newColumns))
  }

  const getCategorySelectionState = (
    category: ColumnCategory,
  ): 'none' | 'partial' | 'full' => {
    const categoryColumns = groupedColumns.get(category) || []
    const userToggleableColumns = categoryColumns.filter(
      (col) =>
        !col.hidden && col.name !== 'chapter_name' && col.name !== 'name',
    )
    const selectedCount = userToggleableColumns.filter((col) =>
      localColumns.includes(col.name),
    ).length

    if (selectedCount === 0) return 'none'
    if (selectedCount === userToggleableColumns.length) return 'full'
    return 'partial'
  }

  const handleOpenChange = (open: boolean) => {
    if (
      !open &&
      JSON.stringify(localColumns) !== JSON.stringify(visibleColumns)
    ) {
      onColumnsChange(localColumns)
    }
    if (open) {
      setLocalColumns(visibleColumns)
      setSearch('')
    }
    setIsOpen(open)
  }

  return (
    <Dialog open={isOpen} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>
        <Button variant="outline" size="sm" className="h-12">
          <Columns3 className="mr-2 h-4 w-4" />
          Columns ({visibleColumns.length})
        </Button>
      </DialogTrigger>
      <DialogContent className="w-[calc(100vw-2rem)] max-w-5xl h-[calc(100vh-4rem)] max-h-[48rem] flex flex-col">
        <DialogHeader>
          <DialogTitle>Select Columns</DialogTitle>
        </DialogHeader>

        {/* Search */}
        <div className="relative">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Search columns..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-9"
          />
        </div>

        {/* Column groups in multi-column grid */}
        <div className="flex-1 overflow-y-auto min-h-0">
          {filteredGroups.size === 0 ? (
            <div className="flex items-center justify-center py-12 text-muted-foreground">
              No columns matching &ldquo;{search}&rdquo;
            </div>
          ) : (
            <div className="columns-1 sm:columns-2 lg:columns-3 gap-6 py-1">
              {Array.from(filteredGroups.entries()).map(
                ([category, columns]) => {
                  const selectionState = getCategorySelectionState(category)
                  const categorySlug = slugifyCategory(category)

                  return (
                    <div
                      key={category}
                      className="break-inside-avoid mb-5 space-y-2"
                    >
                      {isSearching ? (
                        <span className="font-medium text-primary">
                          {category}
                        </span>
                      ) : (
                        <div className="flex items-center gap-2">
                          <Checkbox
                            id={`category-${categorySlug}`}
                            checked={
                              selectionState === 'full'
                                ? true
                                : selectionState === 'partial'
                                  ? 'indeterminate'
                                  : false
                            }
                            onCheckedChange={() =>
                              handleToggleCategory(category)
                            }
                          />
                          <Label
                            htmlFor={`category-${categorySlug}`}
                            className="cursor-pointer font-medium text-primary"
                          >
                            {category}
                          </Label>
                        </div>
                      )}

                      <div className="ml-6 space-y-1.5">
                        {columns
                          .filter(
                            (col) =>
                              !col.hidden &&
                              (col.name !== 'chapter_name' ||
                                isChapterColumnShown),
                          )
                          .map((col) => {
                            const isNameColumn = col.name === 'name'
                            const isChapterColumn = col.name === 'chapter_name'
                            const isDisabled = isNameColumn || isChapterColumn

                            return (
                              <div
                                key={col.name}
                                className="flex items-center gap-2"
                              >
                                <Checkbox
                                  id={`column-${col.name}`}
                                  checked={localColumns.includes(col.name)}
                                  onCheckedChange={() =>
                                    handleToggleColumn(col.name)
                                  }
                                  disabled={isDisabled}
                                />
                                <div
                                  className={`inline-flex items-center gap-1 text-sm ${isDisabled ? 'opacity-60' : ''}`}
                                >
                                  <Label
                                    htmlFor={`column-${col.name}`}
                                    className={
                                      isDisabled
                                        ? 'cursor-default'
                                        : 'cursor-pointer'
                                    }
                                  >
                                    {col.label}
                                  </Label>
                                  {col.description ? (
                                    <Popover>
                                      <PopoverTrigger asChild>
                                        <button
                                          type="button"
                                          aria-label={`About ${col.label}`}
                                          onClick={(e) => e.stopPropagation()}
                                          className="inline-flex items-center"
                                        >
                                          <CircleHelp className="h-3.5 w-3.5 text-muted-foreground" />
                                        </button>
                                      </PopoverTrigger>
                                      <PopoverContent
                                        className="w-64 p-2 text-xs"
                                        side="top"
                                        align="start"
                                      >
                                        {col.description}
                                      </PopoverContent>
                                    </Popover>
                                  ) : null}
                                  {isNameColumn && (
                                    <span className="ml-1 text-xs text-muted-foreground">
                                      (required)
                                    </span>
                                  )}
                                  {isChapterColumn && isChapterColumnShown && (
                                    <span className="ml-1 text-xs text-muted-foreground">
                                      (auto)
                                    </span>
                                  )}
                                </div>
                              </div>
                            )
                          })}
                      </div>
                    </div>
                  )
                },
              )}
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="flex justify-end border-t pt-4">
          <Button onClick={() => handleOpenChange(false)}>Done</Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}
