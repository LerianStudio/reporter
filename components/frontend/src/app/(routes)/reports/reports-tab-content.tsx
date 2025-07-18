'use client'

import React, { useState, useEffect } from 'react'
import { useIntl } from 'react-intl'
import { LayoutGridIcon, Plus, TableIcon } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useListReports } from '@/client/reports'
import { useQueryParams } from '@/hooks/use-query-params'
import { useOrganization } from '@lerianstudio/console-layout'
import { EntityBox } from '@/components/entity-box'
import { ReportsSheet } from './reports-sheet'
import { ReportsCardGrid } from './reports-card-grid'
import { PaginationLimitField } from '@/components/form/pagination-limit-field'
import { FormProvider } from 'react-hook-form'
import { ReportsDataTable } from './reports-data-table'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger
} from '@/components/ui/tooltip'

export function ReportsTabContent() {
  const intl = useIntl()
  const [total, setTotal] = useState(0)
  const [isSheetOpen, setIsSheetOpen] = useState(false)
  const { currentOrganization } = useOrganization()

  const [mode, setMode] = useState<'grid' | 'table'>('grid')

  const { form, searchValues, pagination } = useQueryParams({
    total
  })

  // Fetch reports data using searchValues from useQueryParams
  const {
    data: reportsData,
    isLoading,
    refetch
  } = useListReports({
    ...searchValues,
    organizationId: currentOrganization?.id || ''
  } as any)

  // Download functionality is now handled internally by ReportCard

  // Update total when data changes to manage pagination properly
  useEffect(() => {
    if (!reportsData?.items) {
      setTotal(0)
      return
    }

    // If we have a full page of items, suggest there might be more
    if (reportsData.items.length >= Number(searchValues.limit)) {
      setTotal(Number(searchValues.limit) + 1)
      return
    }

    setTotal(reportsData.items.length)
  }, [reportsData?.items, searchValues.limit])

  // Event handlers
  const handleCreateReport = () => {
    setIsSheetOpen(true)
  }

  return (
    <FormProvider {...form}>
      <>
        <EntityBox.Collapsible className="mb-5">
          <EntityBox.Banner>
            <EntityBox.Header
              title={intl.formatMessage({
                id: 'reports.title',
                defaultMessage: 'Reports'
              })}
              subtitle={intl.formatMessage({
                id: 'reports.subtitle',
                defaultMessage:
                  'Create and manage reports, the output of data processing through templates.'
              })}
              tooltip={intl.formatMessage({
                id: 'reports.tooltip',
                defaultMessage:
                  'Reports are generated documents that transform your ledger data using templates. You can create reports for compliance, analytics, or data export purposes.'
              })}
              tooltipWidth="655px"
            />

            <EntityBox.Actions>
              <EntityBox.CollapsibleTrigger />
              <Button
                icon={<Plus />}
                iconPlacement="end"
                onClick={handleCreateReport}
              >
                {intl.formatMessage({
                  id: 'reports.actions.newReport',
                  defaultMessage: 'New Report'
                })}
              </Button>
            </EntityBox.Actions>
          </EntityBox.Banner>
          <EntityBox.CollapsibleContent>
            <div className="col-start-3 flex items-center justify-end gap-2">
              <PaginationLimitField control={form.control} />
              <TooltipProvider>
                <Tooltip delayDuration={500}>
                  <TooltipTrigger asChild>
                    <Button
                      variant="secondary"
                      className="h-[34px] w-[34px] p-2"
                      onClick={() =>
                        setMode(mode === 'grid' ? 'table' : 'grid')
                      }
                    >
                      {mode === 'grid' ? <LayoutGridIcon /> : <TableIcon />}
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>
                    {mode === 'grid'
                      ? intl.formatMessage({
                          id: 'reports.actions.viewModeTable',
                          defaultMessage: 'Switch to table view'
                        })
                      : intl.formatMessage({
                          id: 'reports.actions.viewModeGrid',
                          defaultMessage: 'Switch to grid view'
                        })}
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </div>
          </EntityBox.CollapsibleContent>
        </EntityBox.Collapsible>

        {mode === 'grid' && (
          <ReportsCardGrid
            reports={reportsData}
            isLoading={isLoading}
            total={total}
            pagination={pagination}
            onCreateReport={handleCreateReport}
          />
        )}

        {mode === 'table' && (
          <ReportsDataTable
            form={form}
            reports={reportsData}
            isLoading={isLoading}
            total={total}
            pagination={pagination}
            onCreateReport={handleCreateReport}
          />
        )}

        {/* Reports Sheet for New Report */}
        <ReportsSheet
          open={isSheetOpen}
          onOpenChange={setIsSheetOpen}
          onSuccess={() => refetch()}
        />
      </>
    </FormProvider>
  )
}
