'use client'

import { AuthedPageContext } from '@/app/authed-page-provider'
import { useContext } from 'react'

export const useAuthedPageContext = () => {
  const ctx = useContext(AuthedPageContext)
  if (!ctx) {
    throw new Error(
      'useAuthedPageContext must be used within an AuthedPageContext provider',
    )
  }
  return ctx
}
