import { ReactNode } from 'react'
import { fetchSession } from 'app/session'
import { redirect } from 'next/navigation'
import { getCookies } from 'lib/auth'

export const AuthedPageLayout = async (props: { children: ReactNode }) => {
  const session = await fetchSession(await getCookies())
  if (!session.user) {
    // Go one level up from the Next.js app root (`/v2`), to get to the
    // absolute URL root, and go to /login from there.
    redirect('/../login')
  }

  return props.children
}
