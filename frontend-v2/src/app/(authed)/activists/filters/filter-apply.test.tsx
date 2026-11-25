import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { IntRangeFilter } from './int-range-filter'

function setup() {
  const onChange = vi.fn()
  const user = userEvent.setup()
  render(<IntRangeFilter label="Events" onChange={onChange} defaultOpen />)
  return { onChange, user }
}

describe('filter Apply behavior', () => {
  it('commits only when Apply is clicked', async () => {
    const { onChange, user } = setup()
    await user.type(screen.getByPlaceholderText('No minimum'), '5')
    expect(onChange).not.toHaveBeenCalled()

    await user.click(screen.getByRole('button', { name: 'Apply' }))
    expect(onChange).toHaveBeenCalledWith({ gte: 5, lt: undefined })
  })

  it('commits on Enter in an input', async () => {
    const { onChange, user } = setup()
    await user.type(screen.getByPlaceholderText('No minimum'), '5{Enter}')
    expect(onChange).toHaveBeenCalledWith({ gte: 5, lt: undefined })
  })

  it('discards the draft when closed without applying', async () => {
    const { onChange, user } = setup()
    await user.type(screen.getByPlaceholderText('No minimum'), '5')
    await user.keyboard('{Escape}')
    expect(onChange).not.toHaveBeenCalled()
  })

  it('keeps Clear visible, disabled until the draft has a value', async () => {
    const { user } = setup()
    const clear = screen.getByRole('button', { name: 'Clear' })
    expect(clear).toBeDisabled()

    await user.type(screen.getByPlaceholderText('No minimum'), '5')
    expect(clear).toBeEnabled()

    await user.click(clear)
    expect(clear).toBeDisabled()
  })
})
