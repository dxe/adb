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
  COLUMN_DEFINITIONS,
  ColumnCategory,
  sortColumnsByDefinitionOrder,
} from './column-definitions'

interface ColumnSelectorProps {
  visibleColumns: ActivistColumnName[]
  onColumnsChange: (columns: ActivistColumnName[]) => void
  showAllChapters: boolean
}

export function ColumnSelector({
  visibleColumns,
  onColumnsChange,
  showAllChapters,
}: ColumnSelectorProps) {
  const [isOpen, setIsOpen] = useState(false)
  const groupedColumns = groupColumnsByCategory()

  const handleToggleColumn = (columnName: ActivistColumnName) => {
    // Prevent removing 'name' column
    if (columnName === 'name' && visibleColumns.includes(columnName)) {
      return
    }

    const newColumns = visibleColumns.includes(columnName)
      ? visibleColumns.filter((col) => col !== columnName)
      : [...visibleColumns, columnName]
    onColumnsChange(sortColumnsByDefinitionOrder(newColumns))
  }

  const handleToggleCategory = (category: ColumnCategory) => {
    const categoryColumns = groupedColumns.get(category) || []
    const categoryColumnNames = categoryColumns.map((col) => col.name)
    const allVisible = categoryColumnNames.every((col) =>
      visibleColumns.includes(col),
    )

    let newColumns: ActivistColumnName[]
    if (allVisible) {
      // Remove all columns in this category
      newColumns = visibleColumns.filter((col) => !categoryColumnNames.includes(col))
    } else {
      // Add all columns in this category
      newColumns = [...visibleColumns]
      categoryColumnNames.forEach((col) => {
        if (!newColumns.includes(col)) {
          newColumns.push(col)
        }
      })
    }
    onColumnsChange(sortColumnsByDefinitionOrder(newColumns))
  }

  const isCategoryFullySelected = (category: ColumnCategory) => {
    const categoryColumns = groupedColumns.get(category) || []
    return categoryColumns.every((col) => visibleColumns.includes(col.name))
  }

  const isCategoryPartiallySelected = (category: ColumnCategory) => {
    const categoryColumns = groupedColumns.get(category) || []
    const selectedCount = categoryColumns.filter((col) =>
      visibleColumns.includes(col.name),
    ).length
    return selectedCount > 0 && selectedCount < categoryColumns.length
  }

  return (
    <Popover open={isOpen} onOpenChange={setIsOpen}>
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
              onClick={() => setIsOpen(false)}
            >
              Done
            </Button>
          </div>

          {/* Chapter name note */}
          {showAllChapters && (
            <div className="rounded-md bg-muted p-2 text-xs text-muted-foreground">
              Chapter name is automatically included when viewing all chapters
            </div>
          )}

          {/* Column groups */}
          {Array.from(groupedColumns.entries()).map(([category, columns]) => {
            const fullySelected = isCategoryFullySelected(category)
            const partiallySelected = isCategoryPartiallySelected(category)

            return (
              <div key={category} className="space-y-2">
                <div className="flex items-center gap-2">
                  <Checkbox
                    id={`category-${category}`}
                    checked={fullySelected}
                    onCheckedChange={() => handleToggleCategory(category)}
                    className={
                      partiallySelected && !fullySelected
                        ? 'data-[state=checked]:bg-primary/50'
                        : ''
                    }
                  />
                  <Label
                    htmlFor={`category-${category}`}
                    className="cursor-pointer font-medium text-primary"
                  >
                    {category}
                  </Label>
                </div>

                <div className="ml-6 space-y-1.5">
                  {columns.map((col) => {
                    const isNameColumn = col.name === 'name'
                    return (
                      <div key={col.name} className="flex items-center gap-2">
                        <Checkbox
                          id={`column-${col.name}`}
                          checked={visibleColumns.includes(col.name)}
                          onCheckedChange={() => handleToggleColumn(col.name)}
                          disabled={isNameColumn}
                        />
                        <Label
                          htmlFor={`column-${col.name}`}
                          className={`text-sm ${isNameColumn ? 'cursor-default opacity-60' : 'cursor-pointer'}`}
                        >
                          {col.label}
                          {isNameColumn && (
                            <span className="ml-1 text-xs text-muted-foreground">
                              (required)
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
