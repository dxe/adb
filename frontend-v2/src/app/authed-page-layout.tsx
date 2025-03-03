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
    // Go one level up from the Next.js app root (`/v2`), to get to the
    // absolute URL root, and go to /login from there.
    redirect('/../login')
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
