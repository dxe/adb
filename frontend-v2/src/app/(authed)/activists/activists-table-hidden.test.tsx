import { cleanup, render, within } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import type { ActivistJSON } from '@/lib/api'
import { ActivistTable } from './activists-table'

const SAMPLE_ACTIVISTS: ActivistJSON[] = [
  { id: 1, name: 'Alice', hidden: true },
  { id: 2, name: 'Bob', hidden: false },
]

const SHADED_ROW_CLASS = 'bg-muted/60'

afterEach(cleanup)

function renderTable() {
  const result = render(
    <ActivistTable
      activists={SAMPLE_ACTIVISTS}
      visibleColumns={['name']}
      sort={[]}
      onSortChange={() => {}}
      onActivistClick={() => {}}
    />,
  )
  return {
    ...result,
    table: within(within(result.container).getByTestId('activists-table')),
  }
}

describe('ActivistTable hidden activists', () => {
  it('shades the desktop row of a hidden activist only', () => {
    const { table } = renderTable()

    // getAllByRole('row') includes the header row.
    const [, hiddenRow, visibleRow] = table.getAllByRole('row')

    expect(hiddenRow.classList.contains(SHADED_ROW_CLASS)).toBe(true)
    expect(visibleRow.classList.contains(SHADED_ROW_CLASS)).toBe(false)
  })

  it('shades the mobile card of a hidden activist only', () => {
    const { container } = renderTable()
    const cards = within(container)

    expect(
      cards.getByTestId('activist-card-1').classList.contains(SHADED_ROW_CLASS),
    ).toBe(true)
    expect(
      cards.getByTestId('activist-card-2').classList.contains(SHADED_ROW_CLASS),
    ).toBe(false)
  })
})
