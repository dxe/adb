import { PropsWithChildren } from 'react'
import { redirect } from 'next/navigation'
import { getCachedSession } from '@/app/session'
import { AuthedPageProvider } from '@/app/authed-page-provider'
import { Navbar } from '@/components/nav'

export default async function AuthedLayout({ children }: PropsWithChildren) {
  const session = await getCachedSession()
  if (!session.user) redirect('/login')

  return (
    <AuthedPageProvider ctx={{ user: session.user }}>
      <Navbar />
      {children}
    </AuthedPageProvider>
  )
}
