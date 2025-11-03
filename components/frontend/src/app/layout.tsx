/* eslint-disable @next/next/no-sync-scripts */

import 'reflect-metadata'
import React from 'react'
import { Inter } from 'next/font/google'
import { Metadata } from 'next'
import { getMetadata } from '../../services/configs/application-config'
import App from './app'
import { getRuntimeEnv } from '@lerianstudio/console-layout'

const inter = Inter({ subsets: ['latin'] })

export default async function RootLayout({
  children
}: {
  children: React.ReactNode
}) {
  const basePath =
    getRuntimeEnv('NEXT_PUBLIC_REPORTER_UI_BASE_PATH') ??
    process.env.NEXT_PUBLIC_REPORTER_UI_BASE_PATH

  return (
    <html suppressHydrationWarning>
      <head>
        <script src={`${basePath}/runtime-env.js`} />
      </head>
      <body suppressHydrationWarning className={inter.className}>
        <App>{children}</App>
      </body>
    </html>
  )
}

export async function generateMetadata(props: {}): Promise<Metadata> {
  const { title, icons, description } = await getMetadata()

  return {
    title: title,
    icons: icons,
    description: description,
    ...(await props)
  }
}
