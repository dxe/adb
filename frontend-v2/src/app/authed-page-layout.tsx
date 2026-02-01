import { PropsWithChildren } from 'react'
import { AuthedPageProvider } from './authed-page-provider'
import { getServerUser } from '@/lib/server-user'

export const AuthedPageLayout = async ({
  pageName,
  children,
}: PropsWithChildren<{
  /** The name of the active page, corresponding to the name in Vue. */
  pageName: string
}>) => {
  const user = await getServerUser()

  return (
    <AuthedPageProvider
      ctx={{
        pageName,
        user,
      }}
    >
      {children}
    </AuthedPageProvider>
  )
}
