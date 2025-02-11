import { ReactNode } from 'react'
import { fetchSession } from 'app/session'
import { redirect } from 'next/navigation'

export const AuthedPageLayout = async (props: { children: ReactNode }) => {
  const session = await fetchSession()
  if (!session.user) {
    redirect('/../login')
  }

  return props.children
}
