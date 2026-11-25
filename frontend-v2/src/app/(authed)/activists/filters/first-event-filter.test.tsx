import { render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { FirstEventFilter, formatFirstEvent } from './first-event-filter'
import type { FirstEventFilterValue } from '../filter-types'

const NULL_LABEL = 'Include activists with no events'

function renderFilter(value?: FirstEventFilterValue) {
  return render(
    <FirstEventFilter value={value} onChange={vi.fn()} defaultOpen />,
  )
}

describe('FirstEventFilter event types', () => {
  it('hides the type checkboxes until a mode is picked', () => {
    renderFilter()
    expect(screen.getByLabelText('First event type is')).toHaveTextContent(
      /^\s*$/,
    )
    expect(screen.queryByLabelText('Outreach')).toBeNull()
  })

  it('shows the mode and checked types', () => {
    renderFilter({ types: { mode: 'exclude', values: ['Outreach'] } })
    expect(screen.getByLabelText('First event type is')).toHaveTextContent(
      'not one of',
    )
    expect(screen.getByLabelText('Outreach')).toBeChecked()
    expect(screen.getByLabelText('Meeting')).not.toBeChecked()
  })
})

describe('FirstEventFilter null option', () => {
  it('is disabled with no date bounds or types', () => {
    renderFilter()
    expect(screen.getByLabelText(NULL_LABEL)).toBeDisabled()
  })

  it('is enabled with one date bound', () => {
    renderFilter({ date: { gte: { mode: 'relative', daysOffset: -30 } } })
    expect(screen.getByLabelText(NULL_LABEL)).toBeEnabled()
  })

  it('is enabled with types and no date bounds', () => {
    renderFilter({
      types: { mode: 'include', values: ['Outreach'] },
      orNull: true,
    })
    expect(screen.getByLabelText(NULL_LABEL)).toBeEnabled()
    expect(screen.getByLabelText(NULL_LABEL)).toBeChecked()
  })

  it('is disabled with both date bounds, even with types', () => {
    renderFilter({
      date: {
        gte: { mode: 'relative', daysOffset: -30 },
        lt: { mode: 'relative', daysOffset: 0 },
      },
      types: { mode: 'include', values: ['Outreach'] },
    })
    expect(screen.getByLabelText(NULL_LABEL)).toBeDisabled()
  })
})

describe('formatFirstEvent', () => {
  it('combines types and date range', () => {
    expect(
      formatFirstEvent({
        date: { lt: { mode: 'relative', daysOffset: -30 } },
        types: { mode: 'include', values: ['Outreach', 'Action'] },
      }),
    ).toBe('Outreach, Action · Over 1 month ago')
  })

  it('describes excluded types and "or none"', () => {
    expect(
      formatFirstEvent({
        types: { mode: 'exclude', values: ['Meeting'] },
        orNull: true,
      }),
    ).toBe('not Meeting or none')
  })

  it('handles date only', () => {
    expect(
      formatFirstEvent({ date: { lt: { mode: 'relative', daysOffset: 0 } } }),
    ).toBe('Before today')
  })
})
