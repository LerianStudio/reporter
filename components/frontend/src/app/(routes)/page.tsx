'use client'

import React from 'react'
import { useIntl } from 'react-intl'
import { HelpCircle, ExternalLink } from 'lucide-react'
import { OverviewTabContent } from './overview/overview-tab-content'
import { TemplatesTabContent } from './templates/templates-tab-content'
import { ReportsTabContent } from './reports/reports-tab-content'
import { useTabs } from '@/hooks/use-tabs'
import { useOrganization } from '@lerianstudio/console-layout'
import { env } from '@/lib/env'
import {
  Alert,
  AlertDescription,
  AlertTitle,
  ApplicationBreadcrumb,
  Button,
  getBreadcrumbPaths,
  PageHeader,
  PageHeaderInfoTitle,
  PageHeaderWrapper,
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger
} from '@lerianstudio/sindarian-ui'

export default function Page() {
  const intl = useIntl()
  const { currentOrganization } = useOrganization()

  const { activeTab, handleTabChange } = useTabs({
    initialValue: 'overview'
  })

  return (
    <>
      <ApplicationBreadcrumb
        paths={getBreadcrumbPaths([
          {
            name: currentOrganization.legalName,
            href: `/organizations/${currentOrganization.id}`
          },
          {
            name: intl.formatMessage({
              id: 'smartTemplates.breadcrumb.smartTemplates',
              defaultMessage: 'Reporter'
            }),
            href: `/`
          },
          {
            name: intl.formatMessage({
              id: 'smartTemplates.breadcrumb.overview',
              defaultMessage: 'Overview'
            }),
            active: () => activeTab === 'overview'
          },
          {
            name: intl.formatMessage({
              id: 'smartTemplates.breadcrumb.templates',
              defaultMessage: 'Templates'
            }),
            active: () => activeTab === 'templates'
          },
          {
            name: intl.formatMessage({
              id: 'smartTemplates.breadcrumb.reports',
              defaultMessage: 'Reports'
            }),
            active: () => activeTab === 'reports'
          }
        ])}
      />

      <PageHeader>
        <PageHeaderWrapper className="border-none">
          <PageHeaderInfoTitle
            title={intl.formatMessage({
              id: 'smartTemplates.title',
              defaultMessage: 'Reporter'
            })}
          />
        </PageHeaderWrapper>

        <Alert className="relative">
          <div className="absolute top-4 left-4">
            <HelpCircle className="h-6 w-6 text-zinc-600" />
          </div>

          <div className="pr-8 pl-10">
            <div className="mb-4 space-y-2">
              <AlertTitle className="text-sm font-medium text-zinc-600">
                {intl.formatMessage({
                  id: 'smartTemplates.about.title',
                  defaultMessage: 'About Reporter'
                })}
              </AlertTitle>
              <AlertDescription className="text-sm leading-relaxed font-medium text-zinc-500">
                {intl.formatMessage({
                  id: 'smartTemplates.about.description',
                  defaultMessage:
                    'Generate dynamic, data-driven reports using plain-text templates (.tpl). Reporter use simple placeholders to pull data directly from the database and renders reports in CSV, XML, HTML, or TXT, always matching the structure defined in the original file.'
                })}
              </AlertDescription>
            </div>

            <div className="flex items-center gap-6">
              <Button
                variant="ghost"
                size="sm"
                className="h-auto p-0 text-sm font-medium text-zinc-800 hover:bg-transparent hover:text-zinc-900"
                icon={<ExternalLink className="h-4 w-4" />}
                iconPlacement="end"
                onClick={() => {
                  window.open(env.DOCS_URL, '_blank', 'noopener,noreferrer')
                }}
              >
                {intl.formatMessage({
                  id: 'common.actions.readDocs',
                  defaultMessage: 'Read the Docs'
                })}
              </Button>
            </div>
          </div>
        </Alert>
      </PageHeader>

      <Tabs
        defaultValue="overview"
        className="w-full"
        value={activeTab}
        onValueChange={handleTabChange}
      >
        <TabsList className="mb-6 grid w-fit grid-cols-3">
          <TabsTrigger value="overview" className="">
            {intl.formatMessage({
              id: 'smartTemplates.tabs.overview',
              defaultMessage: 'Overview'
            })}
          </TabsTrigger>
          <TabsTrigger value="templates">
            {intl.formatMessage({
              id: 'smartTemplates.tabs.templates',
              defaultMessage: 'Templates'
            })}
          </TabsTrigger>
          <TabsTrigger value="reports">
            {intl.formatMessage({
              id: 'smartTemplates.tabs.reports',
              defaultMessage: 'Reports'
            })}
          </TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="mt-0">
          <OverviewTabContent />
        </TabsContent>

        <TabsContent value="templates">
          <TemplatesTabContent />
        </TabsContent>

        <TabsContent value="reports">
          <ReportsTabContent />
        </TabsContent>
      </Tabs>
    </>
  )
}
