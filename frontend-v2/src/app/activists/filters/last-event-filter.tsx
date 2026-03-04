'use client'

import { DateRangeFilter } from './date-range-filter'
import type { DateRangeFilterValue } from '../filter-types'

interface LastEventFilterProps {
  value?: DateRangeFilterValue
  onChange: (value?: DateRangeFilterValue) => void
}

export function LastEventFilter({ value, onChange }: LastEventFilterProps) {
  return (
    <DateRangeFilter
      label="Last event"
      value={value}
      onChange={onChange}
      nullLabel="Include activists with no events"
    />
  )
}
