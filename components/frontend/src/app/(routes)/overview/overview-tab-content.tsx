'use client'

import React from 'react'
import { useIntl } from 'react-intl'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardFooter
} from '@/components/ui/card'
import { useRouter } from 'next/navigation'

type OverviewCardProps = {
  title: string
  description: string
  action: string
  onClick?: () => void
}

const OverviewCard = ({
  title,
  description,
  action,
  onClick
}: OverviewCardProps) => {
  return (
    <Card className="flex flex-col justify-between">
      <CardHeader>
        <CardTitle className={`text-sm font-medium tracking-wide capitalize`}>
          {title}
        </CardTitle>
        <CardDescription className="text-sm leading-relaxed text-zinc-600">
          {description}
        </CardDescription>
      </CardHeader>
      <CardFooter className="p-0">
        <Button
          className="w-full bg-zinc-900 text-white hover:bg-zinc-800"
          size="lg"
          onClick={onClick}
        >
          {action}
        </Button>
      </CardFooter>
    </Card>
  )
}

export function OverviewTabContent() {
  const intl = useIntl()
  const router = useRouter()

  return (
    <div className="grid gap-6 md:grid-cols-1 lg:grid-cols-3">
      {/* Templates Card */}
      <OverviewCard
        title={intl.formatMessage({
          id: 'smartTemplates.cards.templates.title',
          defaultMessage: 'Templates'
        })}
        description={intl.formatMessage({
          id: 'smartTemplates.cards.templates.description',
          defaultMessage:
            "Here are the templates you've created and loaded into the Plugin domain."
        })}
        action={intl.formatMessage({
          id: 'smartTemplates.cards.templates.action',
          defaultMessage: 'Manage Templates'
        })}
        onClick={() => {
          router.push('/?tab=templates')
        }}
      />

      {/* Reports Card */}
      <OverviewCard
        title={intl.formatMessage({
          id: 'smartTemplates.cards.reports.title',
          defaultMessage: 'Reports'
        })}
        description={intl.formatMessage({
          id: 'smartTemplates.cards.reports.description',
          defaultMessage:
            'Here are the reports, the output of data processing through templates.'
        })}
        action={intl.formatMessage({
          id: 'smartTemplates.cards.reports.action',
          defaultMessage: 'Manage Reports'
        })}
        onClick={() => {
          router.push('/?tab=reports')
        }}
      />

      {/* Documentation Card */}
      <OverviewCard
        title={intl.formatMessage({
          id: 'smartTemplates.cards.docs.title',
          defaultMessage: 'Smart Template Docs'
        })}
        description={intl.formatMessage({
          id: 'smartTemplates.cards.docs.description',
          defaultMessage:
            "Need help or additional documentation? We're here to help."
        })}
        action={intl.formatMessage({
          id: 'smartTemplates.cards.docs.action',
          defaultMessage: 'Read the Docs'
        })}
        onClick={() => {
          window.open(
            'https://docs.lerian.studio/docs/smart-templates',
            '_blank',
            'noopener,noreferrer'
          )
        }}
      />
    </div>
  )
}
