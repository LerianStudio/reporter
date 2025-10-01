import React from 'react'
import { EntityDataTable } from '@/components/entity-data-table'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableHeader,
  TableRow
} from '@/components/ui/table'
import { useIntl } from 'react-intl'

type ReportTableSkeletonProps = {
  rowCount?: number
}

export function ReportTableSkeleton({
  rowCount = 5
}: ReportTableSkeletonProps) {
  const intl = useIntl()

  return (
    <EntityDataTable.Root>
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
            {Array.from({ length: rowCount }, (_, index) => (
              <TableRow key={index}>
                <TableCell className="px-6 py-4">
                  <Skeleton className="h-4 w-32" />
                </TableCell>
                <TableCell className="px-6 py-4">
                  <Skeleton className="h-4 w-24" />
                </TableCell>
                <TableCell className="px-6 py-4">
                  <Skeleton className="h-5 w-20 rounded-full" />
                </TableCell>
                <TableCell className="px-6 py-4">
                  <div className="flex justify-center">
                    <Skeleton className="h-5 w-16 rounded-[10px]" />
                  </div>
                </TableCell>
                <TableCell className="px-6 py-4">
                  <Skeleton className="h-4 w-36" />
                </TableCell>
                <TableCell className="w-0 px-6 py-4 text-right">
                  <Skeleton className="h-8 w-8 rounded-md" />
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    </EntityDataTable.Root>
  )
}
