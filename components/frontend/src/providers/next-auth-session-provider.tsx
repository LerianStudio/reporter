'use client'

import React from 'react'
import { SessionProvider } from 'next-auth/react'
import { getRuntimeEnv } from '@lerianstudio/console-layout'

type NextAuthSessionProviderProps = {
  children: React.ReactNode
}

const NextAuthSessionProvider = ({
  children
}: NextAuthSessionProviderProps) => {
  const basePath =
    getRuntimeEnv('NEXT_PUBLIC_MIDAZ_CONSOLE_BASE_URL') ??
    process.env.NEXT_PUBLIC_MIDAZ_CONSOLE_BASE_URL

  return (
    <SessionProvider
      basePath={basePath + '/api/auth'}
      refetchInterval={5 * 60}
      refetchOnWindowFocus={true}
    >
      {children}
    </SessionProvider>
  )
}

export default NextAuthSessionProvider
