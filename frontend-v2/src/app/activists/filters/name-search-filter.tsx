'use client'

import { useState, useEffect, useMemo } from 'react'
import { liteDebounce } from '@tanstack/pacer-lite'
import { Input } from '@/components/ui/input'
import type { FilterState } from '../query-state'

interface NameSearchFilterProps {
  filters: FilterState
  onFiltersChange: (filters: FilterState) => void
}

export function NameSearchFilter({
  filters,
  onFiltersChange,
}: NameSearchFilterProps) {
  const [nameInput, setNameInput] = useState(filters.nameSearch)
  const [prevNameSearch, setPrevNameSearch] = useState(filters.nameSearch)

  // Keep local input in sync when nameSearch changes externally
  // (e.g. URL navigation between preset views).
  if (filters.nameSearch !== prevNameSearch) {
    setPrevNameSearch(filters.nameSearch)
    setNameInput(filters.nameSearch)
  }

  const debouncedNameSearchChange = useMemo(
    () =>
      liteDebounce(
        (nextNameSearch: string, nextFilters: FilterState) => {
          onFiltersChange({ ...nextFilters, nameSearch: nextNameSearch })
        },
        { wait: 300 },
      ),
    [onFiltersChange],
  )

  useEffect(() => {
    if (nameInput === filters.nameSearch) return

    debouncedNameSearchChange(nameInput, filters)
  }, [nameInput, filters, debouncedNameSearchChange])

  return (
    <Input
      type="text"
      placeholder="Search activists by name..."
      value={nameInput}
      onChange={(e) => setNameInput(e.target.value)}
    />
  )
}
