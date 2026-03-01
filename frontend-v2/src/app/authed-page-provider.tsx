'use client'

import { createContext, PropsWithChildren } from 'react'
import { User } from './session'

export type TAuthedPageContext = {
  user: User
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
