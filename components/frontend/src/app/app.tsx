import React from 'react'
import '@/app/globals.css'
import { QueryProvider } from '@/providers/query-provider'
import { ReactQueryDevtools } from '@tanstack/react-query-devtools'
import { LocalizationProvider } from '@/lib/intl'
import { ThemeProvider } from '@/lib/theme'
import ZodSchemaProvider from '@/lib/zod/zod-schema-provider'
import { Toaster } from '@/components/ui/toast/toaster'
import DayjsProvider from '@/providers/dayjs-provider'
import NextAuthSessionProvider from '@/providers/next-auth-session-provider'

export default async function App({ children }: { children: React.ReactNode }) {
  return (
    <NextAuthSessionProvider>
      <LocalizationProvider>
        <QueryProvider>
          <ThemeProvider>
            <DayjsProvider>
              <ZodSchemaProvider>
                {children}
                <Toaster />
              </ZodSchemaProvider>
            </DayjsProvider>
          </ThemeProvider>
          <ReactQueryDevtools initialIsOpen={false} />
        </QueryProvider>
      </LocalizationProvider>
    </NextAuthSessionProvider>
  )
}
