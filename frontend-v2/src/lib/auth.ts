import { headers } from 'next/headers'

export async function getCookies() {
  return (await headers()).get('Cookie') ?? undefined
}
