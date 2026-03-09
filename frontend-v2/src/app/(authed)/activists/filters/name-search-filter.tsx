'use client'

import { Input } from '@/components/ui/input'
import { useDebouncedState } from '@/hooks/use-debounced-state'
import type { FilterState } from '../query-state'

interface NameSearchFilterProps {
  filters: FilterState
  onFiltersChange: (filters: FilterState) => void
}

export function NameSearchFilter({
  filters,
  onFiltersChange,
}: NameSearchFilterProps) {
  const [nameInput, setNameInput] = useDebouncedState(
    filters.nameSearch,
    (nextNameSearch) =>
      onFiltersChange({ ...filters, nameSearch: nextNameSearch }),
  )

  return (
    <Input
      type="text"
      placeholder="Search activists by name..."
      value={nameInput}
      onChange={(e) => setNameInput(e.target.value)}
    />
  )
}
