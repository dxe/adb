import { render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { DateRangeFilter } from './date-range-filter'
import type { DateRangeFilterValue } from '../filter-types'

function renderFilter(value?: DateRangeFilterValue) {
  return render(
    <DateRangeFilter
      label="Last event"
      value={value}
      onChange={vi.fn()}
      defaultOpen
      nullLabel="Include activists with no events"
    />,
  )
}

describe('DateRangeFilter null option', () => {
  it('disables the checkbox when neither bound is set', () => {
    renderFilter()

    expect(
      screen.getByLabelText('Include activists with no events'),
    ).toBeDisabled()
    expect(
      screen.getByTitle('Set exactly one bound to enable this option.'),
    ).toBeInTheDocument()
  })

  it('enables the checkbox when exactly one bound is set', () => {
    renderFilter({
      gte: { mode: 'relative', daysOffset: -30 },
    })

    expect(
      screen.getByLabelText('Include activists with no events'),
    ).toBeEnabled()
    expect(
      screen.queryByTitle('Set exactly one bound to enable this option.'),
    ).toBeNull()
  })

  it('disables the checkbox when both bounds are set', () => {
    renderFilter({
      gte: { mode: 'relative', daysOffset: -30 },
      lt: { mode: 'relative', daysOffset: 0 },
    })

    expect(
      screen.getByLabelText('Include activists with no events'),
    ).toBeDisabled()
    expect(
      screen.getByTitle('Set exactly one bound to enable this option.'),
    ).toBeInTheDocument()
  })
})
