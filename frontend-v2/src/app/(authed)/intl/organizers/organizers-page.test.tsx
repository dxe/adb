import { render, screen, within, cleanup } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { afterEach, describe, expect, it } from 'vitest'
import { CHAPTER_ORGANIZERS_QUERY_KEY } from '@/lib/api'
import OrganizersPage from './organizers-page'

afterEach(cleanup)

const SAMPLE_CHAPTERS = [
  {
    ChapterID: 1,
    Name: 'SF Bay Area',
    Organizers: [
      {
        Name: 'Alice',
        Email: 'alice@example.com',
        Phone: '555-0001',
        Facebook: 'alice.fb',
        Twitter: '@alice',
        Instagram: '@alice_ig',
        Linkedin: 'alice-li',
      },
      {
        Name: 'Bob',
        Email: 'bob@example.com',
        Phone: '555-0002',
        Facebook: '',
        Twitter: '',
        Instagram: '',
        Linkedin: '',
      },
    ],
  },
]

function renderPage() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false, staleTime: Infinity },
    },
  })
  queryClient.setQueryData([...CHAPTER_ORGANIZERS_QUERY_KEY], SAMPLE_CHAPTERS)
  return render(
    <QueryClientProvider client={queryClient}>
      <OrganizersPage />
    </QueryClientProvider>,
  )
}

function getVisibleHeaderLabels() {
  const headers = screen.getAllByRole('columnheader')
  return headers.map(
    (th) => within(th).queryByRole('button')?.getAttribute('aria-label') ?? '',
  )
}

describe('OrganizersPage social media checkbox', () => {
  it('hides social columns by default and shows them when checked', async () => {
    const user = userEvent.setup()
    renderPage()

    let headers = getVisibleHeaderLabels()
    expect(headers).toContain('Chapter')
    expect(headers).toContain('Name')
    expect(headers).not.toContain('Facebook')
    expect(headers).not.toContain('LinkedIn')

    await user.click(screen.getByText('Show social media fields'))

    headers = getVisibleHeaderLabels()
    expect(headers).toContain('Facebook')
    expect(headers).toContain('Twitter')
    expect(headers).toContain('Instagram')
    expect(headers).toContain('LinkedIn')
  })

  it('hides social columns again when unchecked', async () => {
    const user = userEvent.setup()
    renderPage()

    const label = screen.getByText('Show social media fields')
    await user.click(label)
    expect(getVisibleHeaderLabels()).toContain('Facebook')

    await user.click(label)
    expect(getVisibleHeaderLabels()).not.toContain('Facebook')
  })
})

describe('OrganizersPage sorting', () => {
  it('sorts rows descending when clicking a column header twice', async () => {
    const user = userEvent.setup()
    renderPage()

    // Initial order: Alice, Bob (from data)
    let rows = screen.getAllByRole('row').slice(1)
    expect(within(rows[0]).getByText('Alice')).toBeInTheDocument()
    expect(within(rows[1]).getByText('Bob')).toBeInTheDocument()

    // Click Name header twice: asc then desc
    const headerRow = screen.getAllByRole('row')[0]
    const nameButton = within(headerRow).getByRole('button', { name: /Name/ })
    await user.click(nameButton)
    await user.click(nameButton)

    rows = screen.getAllByRole('row').slice(1)
    expect(within(rows[0]).getByText('Bob')).toBeInTheDocument()
    expect(within(rows[1]).getByText('Alice')).toBeInTheDocument()
  })
})
