import { ReactNode } from 'react'
import { fetchSession } from 'app/session'
import { redirect } from 'next/navigation'
import { getCookies } from 'lib/auth'

export const AuthedPageLayout = async (props: { children: ReactNode }) => {
  const session = await fetchSession(await getCookies())
  if (!session.user) {
    redirect('/login')
  }

  return props.children
}
