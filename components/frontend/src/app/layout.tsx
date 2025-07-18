import 'reflect-metadata'
import React from 'react'
import { Inter } from 'next/font/google'
import { Metadata } from 'next'
import { getMetadata } from '../../services/configs/application-config'
import App from './app'

const inter = Inter({ subsets: ['latin'] })

export default async function RootLayout({
  children
}: {
  children: React.ReactNode
}) {
  return (
    <html suppressHydrationWarning>
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
