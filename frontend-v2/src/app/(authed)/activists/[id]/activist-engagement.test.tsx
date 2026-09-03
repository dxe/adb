import { render, screen, within, cleanup } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { afterEach, describe, expect, it, vi } from 'vitest'
import {
  API_PATH,
  apiClient,
  type ActivistTimeline,
  type ActivistTimelineEventItem,
  type ActivistTimelineInteractionItem,
} from '@/lib/api'
import { ActivistEngagement } from './activist-engagement'

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
})

const ACTIVIST_ID = 42

function eventItem(
  overrides: Partial<Omit<ActivistTimelineEventItem, 'event'>> & {
    event?: Partial<ActivistTimelineEventItem['event']>
  } = {},
): ActivistTimelineEventItem {
  const { event, ...rest } = overrides
  return {
    type: 'event',
    id: 1,
    date: '2026-08-20',
    // An event with a recorded start time; has_time false is the other case.
    timestamp: '2026-08-20T18:30:00-07:00',
    has_time: true,
    ...rest,
    event: {
      name: 'Berkeley Chapter Meeting',
      type: 'Community',
      ...event,
    },
  }
}

function interactionItem(
  overrides: Partial<Omit<ActivistTimelineInteractionItem, 'interaction'>> & {
    interaction?: Partial<ActivistTimelineInteractionItem['interaction']>
  } = {},
): ActivistTimelineInteractionItem {
  const { interaction, ...rest } = overrides
  return {
    type: 'interaction',
    id: 1,
    date: '2026-08-10',
    timestamp: '2026-08-10T15:04:00-07:00',
    has_time: true,
    ...rest,
    interaction: {
      method: 'Phone call',
      outcome: 'Interested',
      notes: 'Wants to help with outreach.',
      user_name: 'Sarah Kim',
      ...interaction,
    },
  }
}

function renderEngagement(timeline: ActivistTimeline) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false, staleTime: Infinity } },
  })
  queryClient.setQueryData([API_PATH.ACTIVIST_TIMELINE, ACTIVIST_ID], timeline)
  // Seeded above; a real fetch here would mean the component missed the cache.
  const timelineSpy = vi
    .spyOn(apiClient, 'getActivistTimeline')
    .mockResolvedValue(timeline)
  return {
    queryClient,
    timelineSpy,
    ...render(
      <QueryClientProvider client={queryClient}>
        <ActivistEngagement activistId={ACTIVIST_ID} />
      </QueryClientProvider>,
    ),
  }
}

describe('ActivistEngagement', () => {
  it('renders event and interaction rows', () => {
    renderEngagement({
      items: [eventItem(), interactionItem()],
      truncated: false,
    })

    const [eventRow, interactionRow] = screen.getAllByRole('listitem')

    expect(
      within(eventRow).getByText('Berkeley Chapter Meeting'),
    ).toBeInTheDocument()
    expect(within(eventRow).getByText('Community')).toBeInTheDocument()

    expect(within(interactionRow).getByText('Phone call')).toBeInTheDocument()
    expect(within(interactionRow).getByText('by Sarah Kim')).toBeInTheDocument()
    expect(within(interactionRow).getByText('Interested')).toBeInTheDocument()
    expect(
      within(interactionRow).getByText('Wants to help with outreach.'),
    ).toBeInTheDocument()
  })

  it('shows a time of day only when the server marked one as real', () => {
    renderEngagement({
      items: [
        eventItem({ id: 1, has_time: true }),
        eventItem({ id: 2, has_time: false }),
      ],
      truncated: false,
    })

    const [withTime, withoutTime] = screen.getAllByRole('listitem')
    expect(withTime).toHaveTextContent('6:30 PM')
    // The placeholder instant is noon; it must not surface as a real time.
    expect(withoutTime).not.toHaveTextContent(/\d?\d:\d\d\s?(AM|PM)/)
    expect(withoutTime).toHaveTextContent('Aug 20, 2026')
  })

  it('renders items in the order the server returned them', () => {
    renderEngagement({
      items: [
        interactionItem({
          id: 9,
          date: '2026-08-25',
          timestamp: '2026-08-25T15:04:00-07:00',
          interaction: { method: 'Text message' },
        }),
        eventItem({
          id: 8,
          date: '2026-08-20',
          event: { name: 'Chapter Meeting' },
        }),
        interactionItem({
          id: 7,
          date: '2026-08-01',
          interaction: { method: 'Phone call' },
        }),
      ],
      truncated: false,
    })

    const rows = screen.getAllByRole('listitem')
    expect(rows).toHaveLength(3)
    expect(rows[0]).toHaveTextContent('Text message')
    expect(rows[1]).toHaveTextContent('Chapter Meeting')
    expect(rows[2]).toHaveTextContent('Phone call')
  })

  it('renders the empty state when there is no history', () => {
    renderEngagement({ items: [], truncated: false })

    expect(
      screen.getByText('No events attended or interactions logged yet.'),
    ).toBeInTheDocument()
    expect(screen.queryByRole('listitem')).not.toBeInTheDocument()
  })

  it('notes the item cap only when the timeline is truncated', () => {
    const items = [eventItem(), interactionItem()]

    const { unmount } = renderEngagement({ items, truncated: true })
    expect(
      screen.getByText(/most recent 2 items/, { selector: 'p' }),
    ).toBeInTheDocument()
    unmount()
    cleanup()

    renderEngagement({ items, truncated: false })
    expect(screen.queryByText(/most recent/, { selector: 'p' })).toBeNull()
  })
})
