import { render, screen, within, cleanup } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import {
  afterEach,
  beforeAll,
  beforeEach,
  describe,
  expect,
  it,
  vi,
} from 'vitest'
import { API_PATH, apiClient, type ActivistJSON } from '@/lib/api'
import { Activist } from './activist'

vi.mock('./hide-activist-dialog', () => ({
  HideActivistDialog: () => null,
}))
vi.mock('./merge-activist-dialog', () => ({
  MergeActivistDialog: () => null,
}))

// jsdom doesn't implement these, but Radix's Select and Dialog rely on them for
// their pointer-based interactions.
beforeAll(() => {
  if (!Element.prototype.hasPointerCapture) {
    Element.prototype.hasPointerCapture = () => false
  }
  if (!Element.prototype.releasePointerCapture) {
    Element.prototype.releasePointerCapture = () => {}
  }
  if (!Element.prototype.scrollIntoView) {
    Element.prototype.scrollIntoView = () => {}
  }
})

beforeEach(() => {
  window.sessionStorage.clear()
})

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
})

const ACTIVIST_ID = 42
const ACTIVIST: ActivistJSON = {
  id: ACTIVIST_ID,
  name: 'Test Activist',
  activist_level: 'Organizer',
  email: 'test@example.org',
  phone: '5105550143',
  hiatus: true,
}

function renderDetail() {
  // Saving invalidates the activist query, so the refetch needs to resolve or
  // the page swaps itself out for its error state mid-test.
  vi.spyOn(apiClient, 'getActivist').mockResolvedValue(ACTIVIST)
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false, staleTime: Infinity } },
  })
  queryClient.setQueryData([API_PATH.ACTIVIST_GET, ACTIVIST_ID], ACTIVIST)
  return render(
    <QueryClientProvider client={queryClient}>
      <Activist activistId={ACTIVIST_ID} />
    </QueryClientProvider>,
  )
}

async function openDialog(user: ReturnType<typeof userEvent.setup>) {
  await user.click(screen.getByRole('button', { name: 'Log interaction' }))
  return within(screen.getByRole('dialog'))
}

describe('Log interaction dialog', () => {
  it('opens from the header button with no method preselected', async () => {
    const user = userEvent.setup()
    renderDetail()

    const dialog = await openDialog(user)

    expect(
      dialog.getByRole('heading', { name: 'Log interaction' }),
    ).toBeInTheDocument()
    expect(dialog.getByRole('combobox', { name: 'Method' })).toHaveTextContent(
      'Choose one',
    )
    expect(dialog.getByRole('button', { name: 'Save' })).toBeDisabled()
  })

  it('shows who is being logged, without the page header actions', async () => {
    const user = userEvent.setup()
    renderDetail()

    const dialog = await openDialog(user)

    expect(
      dialog.getByRole('heading', { name: 'Test Activist' }),
    ).toBeInTheDocument()
    expect(dialog.getByText('Organizer')).toBeInTheDocument()
    expect(dialog.getByRole('link', { name: '5105550143' })).toHaveAttribute(
      'href',
      'tel:5105550143',
    )
    expect(
      dialog.getByRole('link', { name: 'test@example.org' }),
    ).toHaveAttribute('href', 'mailto:test@example.org')
    expect(dialog.getByText('Hiatus')).toBeInTheDocument()
    expect(
      dialog.queryByRole('button', { name: 'Activist actions' }),
    ).not.toBeInTheDocument()
  })

  it('offers the messaging-app methods alongside the original three', async () => {
    const user = userEvent.setup()
    renderDetail()

    const dialog = await openDialog(user)
    await user.click(dialog.getByRole('combobox', { name: 'Method' }))

    const listbox = within(screen.getByRole('listbox'))
    for (const method of ['SMS', 'Call', 'Email', 'Signal']) {
      expect(listbox.getByRole('option', { name: method })).toBeInTheDocument()
    }
  })

  it('keeps the follow-up section collapsed until it is opened', async () => {
    const user = userEvent.setup()
    renderDetail()

    const dialog = await openDialog(user)
    const toggle = dialog.getByRole('button', { name: 'Follow-up' })
    expect(toggle).toHaveAttribute('aria-expanded', 'false')

    await user.click(toggle)
    expect(toggle).toHaveAttribute('aria-expanded', 'true')
  })

  it('saves the interaction and remembers the method for the next one', async () => {
    const saveSpy = vi
      .spyOn(apiClient, 'saveInteraction')
      .mockResolvedValue({ status: 'success' })

    const user = userEvent.setup()
    renderDetail()

    let dialog = await openDialog(user)
    await user.click(dialog.getByRole('combobox', { name: 'Method' }))
    await user.click(screen.getByRole('option', { name: 'Signal' }))
    await user.click(dialog.getByRole('combobox', { name: 'Outcome' }))
    await user.click(screen.getByRole('option', { name: 'Could not contact' }))
    await user.type(dialog.getByLabelText('Notes'), 'Not on Signal anymore')
    await user.click(dialog.getByRole('button', { name: 'Save' }))

    expect(saveSpy).toHaveBeenCalledWith({
      activist_id: ACTIVIST_ID,
      method: 'Signal',
      outcome: 'Could not contact',
      notes: 'Not on Signal anymore',
      assign_self: true,
      reset_followup: false,
      set_followup: false,
      followup_days: 3,
    })

    // Reopening starts blank except for the method, which carries over.
    dialog = await openDialog(user)
    expect(dialog.getByRole('combobox', { name: 'Method' })).toHaveTextContent(
      'Signal',
    )
    expect(dialog.getByRole('combobox', { name: 'Outcome' })).toHaveTextContent(
      'Choose one',
    )
    expect(dialog.getByLabelText('Notes')).toHaveValue('')
  })

  it('submits the follow-up days when a follow-up is scheduled', async () => {
    const saveSpy = vi
      .spyOn(apiClient, 'saveInteraction')
      .mockResolvedValue({ status: 'success' })

    const user = userEvent.setup()
    renderDetail()

    const dialog = await openDialog(user)
    await user.click(dialog.getByRole('combobox', { name: 'Method' }))
    await user.click(screen.getByRole('option', { name: 'Call' }))
    await user.click(dialog.getByRole('button', { name: 'Follow-up' }))
    await user.click(dialog.getByLabelText('Follow up in'))
    const days = dialog.getByLabelText('Follow-up days')
    await user.clear(days)
    await user.type(days, '7')
    await user.click(dialog.getByRole('button', { name: 'Save' }))

    expect(saveSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        method: 'Call',
        set_followup: true,
        followup_days: 7,
      }),
    )
  })

  it('drops follow-up choices that are hidden again before saving', async () => {
    const saveSpy = vi
      .spyOn(apiClient, 'saveInteraction')
      .mockResolvedValue({ status: 'success' })

    const user = userEvent.setup()
    renderDetail()

    const dialog = await openDialog(user)
    await user.click(dialog.getByRole('combobox', { name: 'Method' }))
    await user.click(screen.getByRole('option', { name: 'Call' }))

    const toggle = dialog.getByRole('button', { name: 'Follow-up' })
    await user.click(toggle)
    await user.click(dialog.getByLabelText('Clear follow-up date'))
    await user.click(toggle)

    await user.click(dialog.getByRole('button', { name: 'Save' }))

    expect(saveSpy).toHaveBeenCalledWith(
      expect.objectContaining({ reset_followup: false }),
    )
  })
})
