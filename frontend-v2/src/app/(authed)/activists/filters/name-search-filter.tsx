'use client'

import { useState } from 'react'
import { useDebouncedCallback } from '@tanstack/react-pacer'
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

  const debouncedUpdate = useDebouncedCallback(
    (v: string) => onFiltersChange({ ...filters, nameSearch: v }),
    { wait: 300 },
  )

  return (
    <Input
      type="text"
      placeholder="Search activists by name..."
      value={nameInput}
      onChange={(e) => {
        setNameInput(e.target.value)
        debouncedUpdate(e.target.value)
      }}
    />
  )
}
