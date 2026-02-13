'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Label } from '@/components/ui/label'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { Columns3 } from 'lucide-react'
import { ActivistColumnName } from '@/lib/api'
import {
  groupColumnsByCategory,
  ColumnCategory,
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
  const groupedColumns = groupColumnsByCategory()
  const slugifyCategory = (category: ColumnCategory) =>
    category
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, '-')
      .replace(/^-+|-+$/g, '')

  const handleToggleColumn = (columnName: ActivistColumnName) => {
    const newColumns = localColumns.includes(columnName)
      ? localColumns.filter((col) => col !== columnName)
      : [...localColumns, columnName]
    setLocalColumns(normalizeColumns(newColumns))
  }

  const handleToggleCategory = (category: ColumnCategory) => {
    const categoryColumns = groupedColumns.get(category) || []
    const categoryColumnNames = categoryColumns
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
      (col) => col.name !== 'chapter_name' && col.name !== 'name',
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
    }
    setIsOpen(open)
  }

  return (
    <Popover open={isOpen} onOpenChange={handleOpenChange}>
      <PopoverTrigger asChild>
        <Button variant="outline" size="sm" className="h-12">
          <Columns3 className="mr-2 h-4 w-4" />
          Columns ({visibleColumns.length})
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-80 max-h-[32rem] overflow-y-auto">
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h4 className="font-medium">Select Columns</h4>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => handleOpenChange(false)}
            >
              Done
            </Button>
          </div>

          {/* Column groups */}
          {Array.from(groupedColumns.entries()).map(([category, columns]) => {
            const selectionState = getCategorySelectionState(category)
            const categorySlug = slugifyCategory(category)

            return (
              <div key={category} className="space-y-2">
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
                    onCheckedChange={() => handleToggleCategory(category)}
                  />
                  <Label
                    htmlFor={`category-${categorySlug}`}
                    className="cursor-pointer font-medium text-primary"
                  >
                    {category}
                  </Label>
                </div>

                <div className="ml-6 space-y-1.5">
                  {columns
                    .filter(
                      (col) =>
                        col.name !== 'chapter_name' || isChapterColumnShown,
                    )
                    .map((col) => {
                      const isNameColumn = col.name === 'name'
                      const isChapterColumn = col.name === 'chapter_name'
                      const isDisabled = isNameColumn || isChapterColumn

                      return (
                        <div key={col.name} className="flex items-center gap-2">
                          <Checkbox
                            id={`column-${col.name}`}
                            checked={localColumns.includes(col.name)}
                            onCheckedChange={() => handleToggleColumn(col.name)}
                            disabled={isDisabled}
                          />
                          <Label
                            htmlFor={`column-${col.name}`}
                            className={`text-sm ${isDisabled ? 'cursor-default opacity-60' : 'cursor-pointer'}`}
                          >
                            {col.label}
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
                          </Label>
                        </div>
                      )
                    })}
                </div>
              </div>
            )
          })}
        </div>
      </PopoverContent>
    </Popover>
  )
}
