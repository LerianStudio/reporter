'use client'

import React from 'react'
import { SessionProvider } from 'next-auth/react'

type NextAuthSessionProviderProps = {
  children: React.ReactNode
}

const NextAuthSessionProvider = ({
  children
}: NextAuthSessionProviderProps) => {
  return (
    <SessionProvider
      basePath={process.env.NEXT_PUBLIC_MIDAZ_CONSOLE_BASE_URL + '/api/auth'}
      refetchInterval={5 * 60}
      refetchOnWindowFocus={true}
    >
      {children}
    </SessionProvider>
  )
}

export default NextAuthSessionProvider
