import { headers } from 'next/headers'

export async function getAuthCookies() {
  return (await headers()).get('Cookie') ?? undefined
}
