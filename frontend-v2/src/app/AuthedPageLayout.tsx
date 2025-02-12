import { ReactNode } from 'react'
import { fetchSession } from 'app/session'
import { redirect } from 'next/navigation'
import { headers } from 'next/headers'

export const AuthedPageLayout = async (props: { children: ReactNode }) => {
  const session = await fetchSession(
    (await headers()).get('Cookie') ?? undefined,
  )
  if (!session.user) {
    redirect('/../login')
  }

  return props.children
}
