import '@/app/globals.css'
import React from 'react'
import { Metadata } from 'next'
import { AuthRedirect, ConsoleLayout } from '@lerianstudio/console-layout'
import { nextAuthOptions } from '@/core/infrastructure/next-auth/next-auth-provider'

export const metadata: Metadata = {
  title: 'Smart Templates | Midaz Console',
  description: 'Manage smart templates in your system.'
}

export default async function RootLayout({
  children
}: {
  children: React.ReactNode
}) {
  return (
    <AuthRedirect nextAuthOptions={nextAuthOptions}>
      <ConsoleLayout>
        <>{children}</>
      </ConsoleLayout>
    </AuthRedirect>
  )
}
