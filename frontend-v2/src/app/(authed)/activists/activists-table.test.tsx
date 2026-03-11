import { render, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import type { ActivistJSON } from '@/lib/api'
import { COLUMN_DEFINITIONS } from './column-definitions'
import { DEFAULT_SORT } from './query-state'
import { ActivistTable } from './activists-table'

const SAMPLE_ACTIVISTS: ActivistJSON[] = [
  { id: 1, name: 'Alice', email: 'alice@example.com' },
  { id: 2, name: 'Bob', email: 'bob@example.com' },
]

const DEFAULT_COLUMNS: (keyof ActivistJSON)[] = ['name', 'email']

function renderTable(
  sort: Parameters<typeof ActivistTable>[0]['sort'],
  onSortChange = () => {},
) {
  const result = render(
    <ActivistTable
      activists={SAMPLE_ACTIVISTS}
      visibleColumns={DEFAULT_COLUMNS}
      sort={sort}
      onSortChange={onSortChange}
    />,
  )
  return {
    ...result,
    table: within(within(result.container).getByTestId('activists-table')),
  }
}

describe('ActivistTable default sort', () => {
  it('does not show sort indicators when sort is empty (default state)', () => {
    const { table } = renderTable([])

    // The "Name" header button should exist
    const nameHeader = table.getByRole('button', { name: 'Name' })
    expect(nameHeader).toBeInTheDocument()

    // No sort arrow icons should be visible anywhere in the table headers.
    // ArrowUp/ArrowDown from lucide-react render as <svg> elements.
    const headers = table.getAllByRole('columnheader')
    for (const th of headers) {
      expect(th.querySelector('svg')).toBeNull()
    }
  })

  it('shows a sort arrow on the active default sort header when sort is explicitly set', () => {
    const { table } = renderTable(DEFAULT_SORT)
    const activeColumn = DEFAULT_SORT[0].column
    const activeLabel = COLUMN_DEFINITIONS.find(
      (column) => column.name === activeColumn,
    )?.label
    expect(activeLabel).toBeDefined()

    const headers = table.getAllByRole('columnheader')
    const activeHeader = headers.find((header) =>
      within(header).queryByRole('button', {
        name: new RegExp(`^${activeLabel}$`, 'i'),
      }),
    )

    expect(activeHeader).toBeDefined()
    expect(activeHeader?.querySelector('svg')).not.toBeNull()

    for (const header of headers) {
      if (header === activeHeader) continue
      expect(header.querySelector('svg')).toBeNull()
    }
  })

  it('clicking a column header calls onSortChange with ascending sort', async () => {
    const user = userEvent.setup()
    const onSortChange = vi.fn()
    const { table } = renderTable([], onSortChange)

    await user.click(table.getByRole('button', { name: 'Email' }))

    expect(onSortChange).toHaveBeenCalledWith([
      { column: 'email', desc: false },
    ])
  })

  it('clicking the active sort column toggles to descending', async () => {
    const user = userEvent.setup()
    const onSortChange = vi.fn()
    const { table } = renderTable(
      [{ column: 'name', desc: false }],
      onSortChange,
    )

    const nameButton = within(table.getAllByRole('columnheader')[0]).getByRole(
      'button',
    )
    await user.click(nameButton)

    expect(onSortChange).toHaveBeenCalledWith([{ column: 'name', desc: true }])
  })
})
