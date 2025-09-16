'use client'

import React from 'react'
import { useIntl } from 'react-intl'
import { Plus, FileText } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { ReportCard } from '@/app/(routes)/reports/report-card'
import { ReportGridSkeleton } from '@/app/(routes)/reports/report-card-skeleton'
import { Pagination, PaginationProps } from '@/components/pagination'
import { PaginationDto } from '@/core/application/dto/pagination-dto'
import { ReportDto } from '@/core/application/dto/report-dto'

type ReportsCardGridProps = {
  reports: PaginationDto<ReportDto> | undefined
  isLoading: boolean
  total: number
  pagination: PaginationProps
  onCreateReport: () => void
}

export const ReportsCardGrid = ({
  reports,
  isLoading,
  total,
  pagination,
  onCreateReport
}: ReportsCardGridProps) => {
  const intl = useIntl()

  const reportsWithFallback = reports || { items: [] }

  return (
    <div className="rounded-lg">
      <div className="">
        {isLoading ? (
          <ReportGridSkeleton count={6} />
        ) : reportsWithFallback.items.length === 0 ? (
          <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
            <Card className="col-span-full h-[214px] overflow-hidden rounded-lg border border-[#F4F4F5] bg-white shadow-sm md:col-span-1">
              <div className="flex h-full flex-col items-center justify-center gap-4 p-8 text-center">
                <div className="flex h-16 w-16 items-center justify-center rounded-lg bg-gray-50">
                  <FileText className="h-8 w-8 text-gray-400" />
                </div>
                <div className="space-y-2">
                  <h3 className="text-lg font-medium text-gray-700">
                    {intl.formatMessage({
                      id: 'reports.empty.title',
                      defaultMessage: 'No reports found'
                    })}
                  </h3>
                  <Button onClick={onCreateReport} size="sm">
                    <Plus className="mr-2 h-4 w-4" />
                    {intl.formatMessage({
                      id: 'reports.empty.action',
                      defaultMessage: 'New Report'
                    })}
                  </Button>
                </div>
              </div>
            </Card>
          </div>
        ) : (
          <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
            {reportsWithFallback.items.map((report) => (
              <ReportCard
                key={report.id}
                report={report}
                className="w-full max-w-none"
              />
            ))}
          </div>
        )}
      </div>

      <div className="space-y-4 py-4">
        <div className="text-muted-foreground flex items-center justify-between text-sm">
          <div>
            {intl.formatMessage(
              {
                id: 'reports.pagination.info',
                defaultMessage:
                  'Showing {count} {count, plural, one {report} other {reports}}.'
              },
              {
                count: reportsWithFallback.items.length
              }
            )}
          </div>
          <div className="flex items-center gap-2">
            <Pagination total={total} {...pagination} />
          </div>
        </div>
      </div>
    </div>
  )
}
