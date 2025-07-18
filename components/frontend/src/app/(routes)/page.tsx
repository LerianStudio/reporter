'use client'

import React from 'react'
import { useIntl } from 'react-intl'
import { HelpCircle, ExternalLink } from 'lucide-react'
import { Breadcrumb } from '@/components/breadcrumb'
import { PageHeader } from '@/components/page-header'
import { Button } from '@/components/ui/button'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
import { OverviewTabContent } from './overview/overview-tab-content'
import { TemplatesTabContent } from './templates/templates-tab-content'
import { ReportsTabContent } from './reports/reports-tab-content'
import { useTabs } from '@/hooks/use-tabs'
import { useOrganization } from '@lerianstudio/console-layout'
import { CollapsibleContent } from '@/components/ui/collapsible'
import { Alert, AlertTitle, AlertDescription } from '@/components/ui/alert'
import { getBreadcrumbPaths } from '@/components/breadcrumb/get-breadcrumb-paths'

export default function Page() {
  const intl = useIntl()
  const { currentOrganization } = useOrganization()
  const [open, setOpen] = React.useState(false)

  const { activeTab, handleTabChange } = useTabs({
    initialValue: 'overview'
  })

  return (
    <>
      {/* Breadcrumb */}
      <Breadcrumb
        paths={getBreadcrumbPaths([
          {
            name: currentOrganization.legalName,
            href: `/organizations/${currentOrganization.id}`
          },
          {
            name: intl.formatMessage({
              id: 'smartTemplates.breadcrumb.smartTemplates',
              defaultMessage: 'Smart Templates'
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

      {/* Plugin Header */}
      <PageHeader.Root open={open} onOpenChange={setOpen}>
        <PageHeader.Wrapper className="border-none">
          <PageHeader.InfoTitle
            title={intl.formatMessage({
              id: 'smartTemplates.title',
              defaultMessage: 'Smart Templates'
            })}
          />
          <PageHeader.ActionButtons>
            <PageHeader.CollapsibleInfoTrigger
              question={intl.formatMessage({
                id: 'smartTemplates.header.about',
                defaultMessage: 'About'
              })}
            />
          </PageHeader.ActionButtons>
        </PageHeader.Wrapper>

        <CollapsibleContent>
          <Alert className="relative">
            {/* Help Circle Icon */}
            <div className="absolute top-4 left-4">
              <HelpCircle className="h-6 w-6 text-zinc-600" />
            </div>

            {/* Content Area */}
            <div className="pr-8 pl-10">
              {/* Text Section */}
              <div className="mb-4 space-y-2">
                <AlertTitle className="text-sm font-medium text-zinc-600">
                  {intl.formatMessage({
                    id: 'smartTemplates.about.title',
                    defaultMessage: 'About Smart Template'
                  })}
                </AlertTitle>
                <AlertDescription className="text-sm leading-relaxed font-medium text-zinc-500">
                  {intl.formatMessage({
                    id: 'smartTemplates.about.description',
                    defaultMessage:
                      'Generate dynamic, data-driven reports using plain-text templates (.tpl). Smart Templates use simple placeholders to pull data directly from the database and renders reports in CSV, XML, HTML, or TXT, always matching the structure defined in the original file.'
                  })}
                </AlertDescription>
              </div>

              {/* Actions Section */}
              <div className="flex items-center gap-6">
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-auto p-0 text-sm font-medium text-zinc-800 hover:bg-transparent hover:text-zinc-900"
                  icon={<ExternalLink className="h-4 w-4" />}
                  iconPlacement="end"
                  onClick={() => {
                    window.open(
                      'https://docs.lerian.studio/docs/smart-templates',
                      '_blank'
                    )
                  }}
                >
                  {intl.formatMessage({
                    id: 'smartTemplates.about.readDocs',
                    defaultMessage: 'Read the Docs'
                  })}
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-auto p-0 text-sm font-medium text-zinc-800 hover:bg-transparent hover:text-zinc-900"
                  onClick={() => setOpen(false)}
                >
                  {intl.formatMessage({
                    id: 'smartTemplates.about.dismiss',
                    defaultMessage: 'Dismiss'
                  })}
                </Button>
              </div>
            </div>
          </Alert>
        </CollapsibleContent>
      </PageHeader.Root>

      {/* Tabs Navigation */}
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
