import { render, screen, cleanup } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { afterEach, describe, expect, it } from 'vitest'
import { ApplyForm } from '@/app/apply/apply-form'

afterEach(cleanup)

function renderForm() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ApplyForm />
    </QueryClientProvider>,
  )
}

describe('ApplyForm', () => {
  it('walks from the local check through to the application fields', async () => {
    const user = userEvent.setup()
    renderForm()

    expect(
      screen.getByText('Do you live within 100 miles of Berkeley, CA?'),
    ).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Yes' }))
    expect(
      screen.getByRole('button', { name: 'Apply now' }),
    ).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Apply now' }))
    expect(
      screen.getByText('Take direct action for animals'),
    ).toBeInTheDocument()
  })
})
