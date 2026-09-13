import { act, cleanup, fireEvent, render, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import type { ActivistJSON } from '@/lib/api'
import { ActivistTable, type ActivistSelection } from './activists-table'

const SAMPLE_ACTIVISTS: ActivistJSON[] = [
  { id: 1, name: 'Alice', email: 'alice@example.com' },
  { id: 2, name: 'Bob', email: 'bob@example.com' },
]

const DEFAULT_COLUMNS: (keyof ActivistJSON)[] = ['name', 'email']

// This project doesn't enable vitest globals, so RTL's auto-cleanup isn't
// registered for us.
afterEach(cleanup)

// Long-press delay in use-long-press, plus a little slack.
const LONG_PRESS_MS = 500

function renderTable(
  selection?: Partial<ActivistSelection> & { selectedIds?: Set<number> },
  onActivistClick?: (id: number) => void,
) {
  const onToggle = vi.fn()
  const onSetMany = vi.fn()
  const result = render(
    <ActivistTable
      activists={SAMPLE_ACTIVISTS}
      visibleColumns={DEFAULT_COLUMNS}
      sort={[]}
      onSortChange={() => {}}
      onActivistClick={onActivistClick}
      selection={
        selection
          ? {
              selectedIds: selection.selectedIds ?? new Set<number>(),
              onToggle,
              onSetMany,
            }
          : undefined
      }
    />,
  )
  return {
    ...result,
    onToggle,
    onSetMany,
    table: within(within(result.container).getByTestId('activists-table')),
  }
}

describe('ActivistTable selection checkboxes', () => {
  it('renders no checkboxes when no selection is passed', () => {
    const { table } = renderTable()
    expect(table.queryAllByRole('checkbox')).toHaveLength(0)
  })

  it('toggles one activist when its row checkbox is clicked', async () => {
    const user = userEvent.setup()
    const { table, onToggle } = renderTable({})

    await user.click(table.getByRole('checkbox', { name: 'Select Bob' }))

    expect(onToggle).toHaveBeenCalledWith(2)
  })

  it('selects every row from the header checkbox, and deselects them once all are selected', async () => {
    const user = userEvent.setup()
    const none = renderTable({})
    await user.click(
      none.table.getByRole('checkbox', { name: 'Select all activists' }),
    )
    expect(none.onSetMany).toHaveBeenCalledWith([1, 2], true)
    none.unmount()

    const all = renderTable({ selectedIds: new Set([1, 2]) })
    await user.click(
      all.table.getByRole('checkbox', { name: 'Deselect all activists' }),
    )
    expect(all.onSetMany).toHaveBeenCalledWith([1, 2], false)
  })
})

describe('ActivistTable mobile card long press', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('selects a card when it is pressed and held', () => {
    const onActivistClick = vi.fn()
    const { getByTestId, onToggle } = renderTable({}, onActivistClick)
    const card = getByTestId('activist-card-1')

    fireEvent.pointerDown(card, { button: 0, clientX: 10, clientY: 10 })
    act(() => vi.advanceTimersByTime(LONG_PRESS_MS))
    fireEvent.pointerUp(card)

    expect(onToggle).toHaveBeenCalledWith(1)

    // The click the browser fires after the press must not open the activist.
    fireEvent.click(card)
    expect(onActivistClick).not.toHaveBeenCalled()
    expect(onToggle).toHaveBeenCalledTimes(1)
  })

  it('does not select when the press is released early', () => {
    const { getByTestId, onToggle } = renderTable({})
    const card = getByTestId('activist-card-1')

    fireEvent.pointerDown(card, { button: 0, clientX: 10, clientY: 10 })
    act(() => vi.advanceTimersByTime(100))
    fireEvent.pointerUp(card)
    act(() => vi.advanceTimersByTime(LONG_PRESS_MS))

    expect(onToggle).not.toHaveBeenCalled()
  })

  it('does not select when the pointer moves away, as when scrolling', () => {
    const { getByTestId, onToggle } = renderTable({})
    const card = getByTestId('activist-card-1')

    fireEvent.pointerDown(card, { button: 0, clientX: 10, clientY: 10 })
    fireEvent.pointerMove(card, { clientX: 10, clientY: 120 })
    act(() => vi.advanceTimersByTime(LONG_PRESS_MS))

    expect(onToggle).not.toHaveBeenCalled()
  })

  it('taps toggle selection instead of opening an activist once something is selected', () => {
    const onActivistClick = vi.fn()
    const { getByTestId, onToggle } = renderTable(
      { selectedIds: new Set([1]) },
      onActivistClick,
    )

    fireEvent.click(getByTestId('activist-card-2'))

    expect(onToggle).toHaveBeenCalledWith(2)
    expect(onActivistClick).not.toHaveBeenCalled()
  })

  it('taps open the activist when nothing is selected', () => {
    const onActivistClick = vi.fn()
    const { getByTestId, onToggle } = renderTable({}, onActivistClick)

    fireEvent.click(getByTestId('activist-card-2'))

    expect(onActivistClick).toHaveBeenCalledWith(2)
    expect(onToggle).not.toHaveBeenCalled()
  })
})
