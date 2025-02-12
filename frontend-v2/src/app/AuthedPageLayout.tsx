import { ReactNode } from 'react'
import { fetchSession } from 'app/session'
import { redirect } from 'next/navigation'
import { getAuthCookies } from 'lib/auth'

export const AuthedPageLayout = async (props: { children: ReactNode }) => {
  const session = await fetchSession(await getAuthCookies())
  if (!session.user) {
    redirect('/../login')
  }

  return props.children
}
