'use client'

import React from 'react'
import { useIntl } from 'react-intl'
import { Plus, FileText } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { ReportCard } from '@/app/(routes)/reports/report-card'
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

  // Fallback data structure for empty state
  const reportsWithFallback = reports || { items: [] }

  return (
    <div className="rounded-lg">
      {/* Reports Grid Content */}
      <div className="">
        {isLoading ? (
          <div className="flex items-center justify-center py-12">
            <div className="text-muted-foreground text-sm">
              {intl.formatMessage({
                id: 'reports.loading',
                defaultMessage: 'Loading reports...'
              })}
            </div>
          </div>
        ) : reportsWithFallback.items.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-12 text-center">
            <FileText className="text-muted-foreground/50 mb-4 h-12 w-12" />
            <h3 className="text-muted-foreground mb-2 text-lg font-medium">
              {intl.formatMessage({
                id: 'reports.empty.title',
                defaultMessage: 'No reports yet'
              })}
            </h3>
            <p className="text-muted-foreground mb-4 text-sm">
              {intl.formatMessage({
                id: 'reports.empty.description',
                defaultMessage:
                  'Generate your first report from a template to get started.'
              })}
            </p>
            <Button onClick={onCreateReport}>
              <Plus className="mr-2 h-4 w-4" />
              {intl.formatMessage({
                id: 'reports.empty.action',
                defaultMessage: 'New Report'
              })}
            </Button>
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

      {/* Footer with pagination */}
      {reportsWithFallback.items.length > 0 && (
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
      )}
    </div>
  )
}
