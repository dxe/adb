import { PropsWithChildren } from 'react'
import { fetchSession } from '@/app/session'
import { redirect } from 'next/navigation'
import { getCookies } from '@/lib/auth'
import { AuthedPageProvider } from './authed-page-provider'

export const AuthedPageLayout = async ({
  pageName,
  children,
}: PropsWithChildren<{
  /** The name of the active page, corresponding to the name in Vue. */
  pageName: string
}>) => {
  const session = await fetchSession(await getCookies())
  if (!session.user) {
    // This goes to /v2/login
    redirect('/login')
  }

  return (
    <AuthedPageProvider
      ctx={{
        pageName,
        user: session.user,
      }}
    >
      {children}
    </AuthedPageProvider>
  )
}
