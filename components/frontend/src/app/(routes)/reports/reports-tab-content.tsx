'use client'

import React, { useState, useCallback } from 'react'
import { useIntl } from 'react-intl'
import { LayoutGridIcon, Plus, TableIcon, FunnelX, Loader2 } from 'lucide-react'
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
import { InputField, DatePickerField } from '@/components/form'
import { ReportFiltersDto } from '@/core/application/dto/report-dto'

const DEFAULT_TOTAL_COUNT = 100000

export function ReportsTabContent() {
  const intl = useIntl()
  const [total, setTotal] = useState(DEFAULT_TOTAL_COUNT)
  const [isSheetOpen, setIsSheetOpen] = useState(false)
  const { currentOrganization } = useOrganization()

  const [mode, setMode] = useState<'grid' | 'table'>('grid')

  const { form, searchValues, pagination } = useQueryParams<ReportFiltersDto>({
    total,
    initialValues: {
      search: '',
      status: undefined,
      templateId: undefined,
      createdAt: undefined
    }
  })

  const {
    data: reportsData,
    isLoading,
    isFetching,
    refetch
  } = useListReports({
    filters: {
      ...searchValues,
      page: parseInt(searchValues.page) || 1,
      limit: parseInt(searchValues.limit) || 10
    },
    organizationId: currentOrganization?.id || ''
  })

  const handleCreateReport = useCallback(() => {
    setIsSheetOpen(true)
  }, [setIsSheetOpen])

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
              {isFetching && (
                <div className="text-muted-foreground flex items-center gap-2 text-sm">
                  <Loader2 className="h-4 w-4 animate-spin" />
                  <span>
                    {intl.formatMessage({
                      id: 'reports.loading.updating',
                      defaultMessage: 'Updating...'
                    })}
                  </span>
                </div>
              )}
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
            <div className="col-span-2 flex grow flex-col gap-4 sm:flex-row">
              <InputField
                name="search"
                placeholder={intl.formatMessage({
                  id: 'common.searchPlaceholder',
                  defaultMessage: 'Search...'
                })}
                control={form.control}
              />

              <div className="sm:w-auto sm:flex-none">
                <DatePickerField
                  name="createdAt"
                  placeholder={intl.formatMessage({
                    id: 'reports.filters.selectDate',
                    defaultMessage: 'Select date'
                  })}
                  control={form.control}
                />
              </div>
            </div>
            <div className="col-start-3 flex items-center justify-end gap-2">
              <PaginationLimitField control={form.control} />
              <Button
                variant="outline"
                className="h-[34px] w-[34px] bg-white p-2 hover:bg-white"
                onClick={() => {
                  form.reset({
                    search: '',
                    status: undefined,
                    templateId: undefined,
                    createdAt: undefined
                  })
                }}
              >
                <FunnelX size={16} />
              </Button>
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

        <ReportsSheet
          open={isSheetOpen}
          onOpenChange={setIsSheetOpen}
          onSuccess={() => refetch()}
        />
      </>
    </FormProvider>
  )
}
