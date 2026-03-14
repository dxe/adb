'use client'

import { createContext, PropsWithChildren } from 'react'
import { type AuthedUser } from '@/lib/api'

export type TAuthedPageContext = {
  user: AuthedUser
}

export const AuthedPageContext = createContext<TAuthedPageContext | undefined>(
  undefined,
)

export const AuthedPageProvider = ({
  ctx,
  children,
}: PropsWithChildren<{ ctx: TAuthedPageContext }>) => {
  return (
    <AuthedPageContext.Provider value={ctx}>
      {children}
    </AuthedPageContext.Provider>
  )
}
