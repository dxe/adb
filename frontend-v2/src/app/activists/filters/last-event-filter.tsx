'use client'

import { DateRangeFilter } from './date-range-filter'

interface LastEventFilterProps {
  value?: string
  onChange: (value?: string) => void
}

export function LastEventFilter({ value, onChange }: LastEventFilterProps) {
  return <DateRangeFilter label="Last event" value={value} onChange={onChange} />
}
