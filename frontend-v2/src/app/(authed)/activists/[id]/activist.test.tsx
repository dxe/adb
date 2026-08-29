import { render, screen, within, cleanup } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest'
import {
  API_PATH,
  apiClient,
  type ActivistJSON,
  type AssignableUser,
} from '@/lib/api'
import { Activist } from './activist'

// HideActivistDialog and MergeActivistDialog and their dependencies are
// unrelated to the existing tests at this time. Stub them out so this file can
// focus on Activist/section-form.
vi.mock('./hide-activist-dialog', () => ({
  HideActivistDialog: () => null,
}))
vi.mock('./merge-activist-dialog', () => ({
  MergeActivistDialog: () => null,
}))

// jsdom doesn't implement these, but Radix's Select relies on them for its
// pointer-based open/select interactions.
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

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
})

const ACTIVIST_ID = 42

function renderDetail(
  activist: ActivistJSON,
  assignableUsers?: AssignableUser[],
) {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false, staleTime: Infinity },
    },
  })
  queryClient.setQueryData([API_PATH.ACTIVIST_GET, ACTIVIST_ID], activist)
  if (assignableUsers) {
    queryClient.setQueryData([API_PATH.USERS_ASSIGNABLE], assignableUsers)
  }
  return {
    queryClient,
    ...render(
      <QueryClientProvider client={queryClient}>
        <Activist activistId={ACTIVIST_ID} />
      </QueryClientProvider>,
    ),
  }
}

describe('Activist Assigned To field (read-only mode)', () => {
  it('shows the assigned_to_name value, not the raw assigned_to id', () => {
    renderDetail({
      id: ACTIVIST_ID,
      name: 'Test Activist',
      assigned_to: 7,
      assigned_to_name: 'Bob Jones',
    })

    const dt = screen.getByText('Assigned To')
    const row = dt.closest('div')
    expect(row).not.toBeNull()
    expect(within(row!).getByText('Bob Jones')).toBeInTheDocument()
    expect(within(row!).queryByText('7')).not.toBeInTheDocument()

    // Only one "Assigned To" row should render (assigned_to_name itself is
    // hidden on the detail page via hideOnDetailPage).
    expect(screen.getAllByText('Assigned To')).toHaveLength(1)
  })

  it('shows a placeholder when assigned_to_name is absent', () => {
    renderDetail({
      id: ACTIVIST_ID,
      name: 'Test Activist',
    })

    const dt = screen.getByText('Assigned To')
    const row = dt.closest('div')
    expect(within(row!).getByText('—')).toBeInTheDocument()
  })
})

describe('Activist Assigned To field (edit mode)', () => {
  const ASSIGNABLE_USERS: AssignableUser[] = [
    { id: 1, name: 'Alice' },
    { id: 2, name: 'Bob Jones' },
  ]

  it('populates the dropdown with the assignable users plus Unassigned', async () => {
    const user = userEvent.setup()
    renderDetail(
      {
        id: ACTIVIST_ID,
        name: 'Test Activist',
        assigned_to: 1,
        assigned_to_name: 'Alice',
      },
      ASSIGNABLE_USERS,
    )

    await user.click(screen.getByRole('button', { name: 'Edit Prospect Info' }))

    const trigger = screen.getByRole('combobox', { name: 'Assigned To' })
    await user.click(trigger)

    const listbox = screen.getByRole('listbox')
    expect(within(listbox).getByText('Unassigned')).toBeInTheDocument()
    expect(within(listbox).getByText('Alice')).toBeInTheDocument()
    expect(within(listbox).getByText('Bob Jones')).toBeInTheDocument()
  })

  it('saves the newly selected user as assigned_to', async () => {
    const activist: ActivistJSON = {
      id: ACTIVIST_ID,
      name: 'Test Activist',
      assigned_to: 1,
      assigned_to_name: 'Alice',
    }
    const patchSpy = vi.spyOn(apiClient, 'patchActivist').mockResolvedValue({
      ...activist,
      assigned_to: 2,
      assigned_to_name: 'Bob Jones',
    })

    const user = userEvent.setup()
    renderDetail(activist, ASSIGNABLE_USERS)

    await user.click(screen.getByRole('button', { name: 'Edit Prospect Info' }))

    const trigger = screen.getByRole('combobox', { name: 'Assigned To' })
    await user.click(trigger)
    await user.click(screen.getByRole('option', { name: 'Bob Jones' }))

    await user.click(screen.getByRole('button', { name: 'Save' }))

    expect(patchSpy).toHaveBeenCalledWith(ACTIVIST_ID, { assigned_to: 2 })
  })
})
