'use client'

// ssr: false b/c default date filter uses new Date() which differs between
// server (UTC) and client (local tz). next/dynamic with ssr: false must live
// in a Client Component, so this thin wrapper exists for that purpose.
// https://github.com/dxe/adb/pull/314#discussion_r2900328919
import dynamic from 'next/dynamic'
import { Loading } from '@/app/loading'

export const EventsPageLoader = dynamic(() => import('./events-page'), {
  ssr: false,
  loading: () => (
    <>
      <h1 className="text-2xl font-semibold">All Events</h1>
      <Loading inline />
    </>
  ),
})
