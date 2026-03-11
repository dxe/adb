'use client'

import { useEffect, useState } from 'react'
import { useDebouncer } from '@tanstack/react-pacer'
import { Input } from '@/components/ui/input'
import type { FilterState } from '../query-state'

interface NameSearchFilterProps {
  filters: FilterState
  onFiltersChange: (filters: FilterState) => void
}

function DebouncedNameSearchInput({
  filters,
  onFiltersChange,
}: NameSearchFilterProps) {
  const [nameInput, setNameInput] = useState(filters.nameSearch)
  const debouncer = useDebouncer(
    (nextNameSearch: string) =>
      onFiltersChange({ ...filters, nameSearch: nextNameSearch }),
    { wait: 300 },
  )

  useEffect(() => {
    return () => debouncer.cancel()
  }, [debouncer])

  return (
    <Input
      type="text"
      placeholder="Search activists by name..."
      value={nameInput}
      onChange={(e) => {
        const nextValue = e.target.value
        setNameInput(nextValue)
        debouncer.maybeExecute(nextValue)
      }}
    />
  )
}

export function NameSearchFilter(props: NameSearchFilterProps) {
  // Remount on external nameSearch changes so pending debounces are canceled
  // before they can write stale state back into the URL.
  return <DebouncedNameSearchInput key={props.filters.nameSearch} {...props} />
}
