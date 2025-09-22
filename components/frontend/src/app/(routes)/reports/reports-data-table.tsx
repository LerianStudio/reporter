'use client'

import React from 'react'
import { useIntl } from 'react-intl'
import { Loader2, MoreVertical } from 'lucide-react'
import dayjs from 'dayjs'
import { EmptyResource } from '@/components/empty-resource'
import { EntityDataTable } from '@/components/entity-data-table'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger
} from '@/components/ui/dropdown-menu'
import {
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableHeader,
  TableRow
} from '@/components/ui/table'
import { FormProvider, UseFormReturn } from 'react-hook-form'
import { Pagination, PaginationProps } from '@/components/pagination'
import { ReportStatusBadge } from './report-status-badge'
import { ReportDto } from '@/core/application/dto/report-dto'
import { PaginationDto } from '@/core/application/dto/pagination-dto'
import { IdTableCell } from '@/components/table/id-table-cell'
import { useDownloadReport } from '@/client/reports'
import { useToast } from '@/hooks/use-toast'
import { useOrganization } from '@lerianstudio/console-layout'
import { ReportTableSkeleton } from '@/app/(routes)/reports/report-table-skeleton'

type ReportsDataTableProps = {
  reports: PaginationDto<ReportDto> | undefined
  form: UseFormReturn<any>
  total: number
  pagination: PaginationProps
  isLoading?: boolean
  onCreateReport: () => void
}

type ReportRowProps = {
  report: ReportDto
}

const FormatBadge = ({ format }: { format: string }) => {
  return (
    <Badge
      variant="outline"
      className="rounded-[10px] border-[#F4F4F5] bg-white px-2 py-0.5 text-xs font-medium text-[#3F3F46]"
    >
      {format}
    </Badge>
  )
}

const ReportRow = ({ report }: ReportRowProps) => {
  const intl = useIntl()
  const { toast } = useToast()
  const { currentOrganization } = useOrganization()

  const downloadMutation = useDownloadReport({
    organizationId: currentOrganization.id,
    reportId: report.id,
    onSuccess: () => {
      toast({
        title: intl.formatMessage({
          id: 'reports.actions.downloadSuccess',
          defaultMessage: 'Report downloaded successfully'
        }),
        variant: 'success'
      })
    }
  })

  const handleDownloadReport = () => {
    downloadMutation.mutate(report.id)
  }

  return (
    <TableRow key={report.id}>
      <TableCell className="font-normal text-[#6B7280]">
        {report.template?.name ?? report.template?.fileName}
      </TableCell>

      <IdTableCell id={report.id} />

      <TableCell className="py-3.5">
        <ReportStatusBadge status={report.status} />
      </TableCell>

      <TableCell className="py-3.5">
        <div className="flex justify-center">
          <FormatBadge format={report.template?.outputFormat.toUpperCase()!} />
        </div>
      </TableCell>

      <TableCell className="font-normal text-[#6B7280]">
        {report.completedAt
          ? dayjs(report.completedAt).format('L - LT')
          : 'N/A'}
      </TableCell>

      <TableCell className="w-0 text-right">
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button
              variant="secondary"
              className="h-8 w-8 border border-[#D4D4D8] bg-white p-2 shadow-sm hover:bg-[#F4F4F5]"
            >
              <MoreVertical className="h-4 w-4 text-[#52525B]" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            {report.status === 'Finished' && (
              <DropdownMenuItem
                disabled={downloadMutation.isPending}
                onClick={handleDownloadReport}
              >
                {intl.formatMessage({
                  id: 'reports.actions.download',
                  defaultMessage: 'Download Report'
                })}
              </DropdownMenuItem>
            )}
          </DropdownMenuContent>
        </DropdownMenu>
      </TableCell>
    </TableRow>
  )
}

export const ReportsDataTable = ({
  reports,
  form,
  total,
  pagination,
  isLoading,
  onCreateReport
}: ReportsDataTableProps) => {
  const intl = useIntl()

  if (isLoading) {
    return (
      <FormProvider {...form}>
        <ReportTableSkeleton rowCount={5} />
      </FormProvider>
    )
  }

  return (
    <EntityDataTable.Root>
      {!reports?.items || reports.items.length === 0 ? (
        <>
          <EmptyResource
            message={intl.formatMessage({
              id: 'reports.empty.message',
              defaultMessage: "You haven't created any reports yet"
            })}
          >
            <Button onClick={onCreateReport}>
              {intl.formatMessage({
                id: 'reports.actions.newReport',
                defaultMessage: 'New Report'
              })}
            </Button>
          </EmptyResource>
          <EntityDataTable.Footer className="rounded-b-lg border-t border-[#E5E7EB] bg-white px-6 py-3">
            <EntityDataTable.FooterText className="text-sm leading-8 font-medium text-[#A1A1AA] italic">
              {intl.formatMessage(
                {
                  id: 'reports.pagination.info',
                  defaultMessage:
                    'Showing {count} {count, plural, one {report} other {reports}}.'
                },
                {
                  count: reports?.items.length ?? 0
                }
              )}
            </EntityDataTable.FooterText>
            <Pagination total={total} {...pagination} />
          </EntityDataTable.Footer>
        </>
      ) : (
        <>
          <TableContainer>
            <Table className="bg-white">
              <TableHeader>
                <TableRow className="border-b border-[#E5E7EB] hover:bg-transparent">
                  <TableHead className="px-6 py-4 text-sm font-medium text-[#52525B]">
                    {intl.formatMessage({
                      id: 'common.field.name',
                      defaultMessage: 'Name'
                    })}
                  </TableHead>
                  <TableHead className="px-6 py-4 text-sm font-medium text-[#52525B]">
                    {intl.formatMessage({
                      id: 'reports.table.reportId',
                      defaultMessage: 'Report ID'
                    })}
                  </TableHead>
                  <TableHead className="px-6 py-4 text-sm font-medium text-[#52525B]">
                    {intl.formatMessage({
                      id: 'common.status',
                      defaultMessage: 'Status'
                    })}
                  </TableHead>
                  <TableHead className="px-6 py-4 text-center text-sm font-medium text-[#52525B]">
                    {intl.formatMessage({
                      id: 'reports.table.format',
                      defaultMessage: 'Format'
                    })}
                  </TableHead>
                  <TableHead className="px-6 py-4 text-sm font-medium text-[#52525B]">
                    {intl.formatMessage({
                      id: 'reports.table.completedAt',
                      defaultMessage: 'Completed At'
                    })}
                  </TableHead>
                  <TableHead className="w-[90px] px-6 py-4 text-right text-sm font-medium text-[#52525B]">
                    {intl.formatMessage({
                      id: 'common.actions',
                      defaultMessage: 'Actions'
                    })}
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {reports.items.map((report) => (
                  <ReportRow key={report.id} report={report} />
                ))}
              </TableBody>
            </Table>
          </TableContainer>

          <EntityDataTable.Footer className="rounded-b-lg border-t border-[#E5E7EB] bg-white px-6 py-3">
            <EntityDataTable.FooterText className="text-sm leading-8 font-medium text-[#A1A1AA] italic">
              {intl.formatMessage(
                {
                  id: 'reports.pagination.info',
                  defaultMessage:
                    'Showing {count} {count, plural, one {report} other {reports}}.'
                },
                {
                  count: reports.items.length
                }
              )}
            </EntityDataTable.FooterText>
            <Pagination total={total} {...pagination} />
          </EntityDataTable.Footer>
        </>
      )}
    </EntityDataTable.Root>
  )
}
