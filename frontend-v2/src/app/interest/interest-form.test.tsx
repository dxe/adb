import { render, screen, cleanup } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { apiClient } from '@/lib/api'
import { InterestForm } from './interest-form'

const mockSearchParams = vi.hoisted(() => ({ current: new URLSearchParams() }))
vi.mock('next/navigation', () => ({
  useSearchParams: () => mockSearchParams.current,
}))

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
})

function renderForm(search = '') {
  mockSearchParams.current = new URLSearchParams(search)
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <InterestForm />
    </QueryClientProvider>,
  )
}

async function fillRequiredFields(user: ReturnType<typeof userEvent.setup>) {
  await user.type(screen.getByLabelText('First Name'), 'Jane')
  await user.type(screen.getByLabelText('Last Name'), 'Doe')
  await user.type(
    screen.getByRole('textbox', { name: 'Email' }),
    'jane@example.com',
  )
  await user.type(screen.getByLabelText('Phone Number'), '5551234567')
  await user.type(screen.getByLabelText('Zip Code'), '94612')
}

describe('InterestForm', () => {
  it('derives the title, description, and referral-apply prefill from URL params', () => {
    renderForm(
      'title=Custom+Title&description=Custom+Description&showReferralApply=true&referralApply=A+friend',
    )

    expect(
      screen.getByRole('heading', { name: 'Custom Title' }),
    ).toBeInTheDocument()
    expect(screen.getByText('Custom Description')).toBeInTheDocument()
    expect(screen.getByLabelText('Who encouraged you to sign up?')).toHaveValue(
      'A friend',
    )
  })

  it('submits the exact values entered by the user, along with the URL-derived chapter and form name', async () => {
    const submitSpy = vi
      .spyOn(apiClient, 'submitInterestForm')
      .mockResolvedValue({ status: 'success' })
    const user = userEvent.setup()
    renderForm(
      'chapterId=7&name=Volunteer&showReferralFriends=true&showReferralApply=true&showReferralOutlet=true',
    )

    await fillRequiredFields(user)
    await user.type(
      screen.getByLabelText(
        'List any existing DxE activists who you are close friends with',
      ),
      'Alex Smith',
    )
    await user.type(
      screen.getByLabelText('Who encouraged you to sign up?'),
      'A friend',
    )
    await user.click(screen.getByRole('radio', { name: 'Social Media' }))
    await user.click(screen.getByRole('checkbox', { name: /Animal Care/ }))
    await user.click(screen.getByRole('button', { name: 'Submit' }))

    expect(submitSpy).toHaveBeenCalledWith({
      chapterId: 7,
      form: 'Volunteer Form',
      name: 'Jane Doe',
      email: 'jane@example.com',
      zip: '94612',
      phone: '5551234567',
      referralFriends: 'Alex Smith',
      referralApply: 'A friend',
      referralOutlet: 'Social Media',
      interests: 'Animal Care',
    })
    expect(
      await screen.findByText('Thank you for your submission!'),
    ).toBeInTheDocument()
  })

  it('joins the checked activism interests into the submitted payload', async () => {
    const submitSpy = vi
      .spyOn(apiClient, 'submitInterestForm')
      .mockResolvedValue({ status: 'success' })
    const user = userEvent.setup()
    renderForm()

    await fillRequiredFields(user)
    await user.click(screen.getByRole('checkbox', { name: /Animal Care/ }))
    await user.click(screen.getByRole('checkbox', { name: /Outreach/ }))
    await user.click(screen.getByRole('button', { name: 'Submit' }))

    expect(submitSpy).toHaveBeenCalledWith(
      expect.objectContaining({ interests: 'Animal Care, Outreach' }),
    )
  })

  it('always submits empty interests for Circle Interest forms, even when interests are checked', async () => {
    const submitSpy = vi
      .spyOn(apiClient, 'submitInterestForm')
      .mockResolvedValue({ status: 'success' })
    const user = userEvent.setup()
    renderForm('name=Circle+Interest')

    await fillRequiredFields(user)
    await user.click(screen.getByRole('checkbox', { name: /Animal Care/ }))
    await user.click(screen.getByRole('button', { name: 'Submit' }))

    expect(submitSpy).toHaveBeenCalledWith(
      expect.objectContaining({ interests: '' }),
    )
  })
})
