import {
  render,
  screen,
  cleanup,
  fireEvent,
  waitFor,
} from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { ApplicationFields } from '@/app/apply/application-fields'

afterEach(cleanup)

describe('ApplicationFields', () => {
  it('submits the parsed payload once all required fields are filled in', async () => {
    const user = userEvent.setup()
    const onSubmit = vi.fn()
    render(<ApplicationFields onSubmit={onSubmit} isSubmitting={false} />)

    await user.type(screen.getByLabelText('First Name'), 'Jane')
    await user.type(screen.getByLabelText('Last Name'), 'Doe')
    await user.type(screen.getByLabelText('Email'), 'jane@example.com')
    await user.type(screen.getByLabelText('Street Address'), '123 Main St')
    await user.type(screen.getByLabelText('City'), 'Berkeley')
    await user.type(screen.getByLabelText('Zip Code'), '94704')
    await user.type(screen.getByLabelText('Phone'), '5551234567')
    fireEvent.change(screen.getByLabelText('Birthday'), {
      target: { value: '2000-01-01' },
    })

    for (const checkbox of screen.getAllByRole('checkbox')) {
      await user.click(checkbox)
    }

    await user.click(screen.getByRole('button', { name: 'Yes' }))
    await user.click(screen.getByRole('button', { name: 'Submit' }))

    await waitFor(() => expect(onSubmit).toHaveBeenCalled())
    expect(onSubmit).toHaveBeenCalledWith(
      expect.objectContaining({
        name: 'Jane Doe',
        firstName: 'Jane',
        lastName: 'Doe',
        email: 'jane@example.com',
        address: '123 Main St',
        city: 'Berkeley',
        zip: '94704',
        phone: '5551234567',
        birthday: '2000-01-01',
        applicationType: 'organizer',
      }),
    )
  })
})
