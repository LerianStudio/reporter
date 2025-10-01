'use client'

import React from 'react'
import { useIntl } from 'react-intl'
import { FileText, Clock, MoreVertical, Download } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardFooter, CardHeader } from '@/components/ui/card'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger
} from '@/components/ui/dropdown-menu'
import { ReportDto } from '@/core/application/dto/report-dto'
import { useDownloadReport } from '@/client/reports'
import { useOrganization } from '@lerianstudio/console-layout'
import { useToast } from '@/hooks/use-toast'
import { ReportStatusBadge } from './report-status-badge'
import dayjs from 'dayjs'
import { cn } from '@/lib/utils'

type DownloadingOverlayProps = {
  isDownloading: boolean
}

function DownloadingOverlay({ isDownloading }: DownloadingOverlayProps) {
  const intl = useIntl()

  if (!isDownloading) {
    return null
  }

  return (
    <div className="absolute inset-0 flex items-center justify-center bg-black/10 backdrop-blur-sm">
      <div className="flex items-center gap-2 rounded-lg bg-white p-3 shadow-lg">
        <div className="h-4 w-4 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
        <span className="text-sm font-medium">
          {intl.formatMessage({
            id: 'reports.downloading',
            defaultMessage: 'Downloading...'
          })}
        </span>
      </div>
    </div>
  )
}

export type ReportCardProps = React.ComponentProps<typeof Card> & {
  report: ReportDto
}

export function ReportCard({ report, className, ...props }: ReportCardProps) {
  const intl = useIntl()
  const { currentOrganization } = useOrganization()
  const { toast } = useToast()

  const downloadMutation = useDownloadReport({
    organizationId: currentOrganization?.id || '',
    reportId: report.id || '',
    onSuccess: () => {
      toast({
        title: intl.formatMessage({
          id: 'reports.download.success.title',
          defaultMessage: 'Download Completed'
        }),
        description: intl.formatMessage({
          id: 'reports.download.success.description',
          defaultMessage: 'Your report has been downloaded successfully.'
        }),
        variant: 'success'
      })
    }
  })

  // Check if report is downloadable
  const isDownloadable = report.status === 'Finished'

  const handleDownload = () => {
    if (isDownloadable && !downloadMutation.isPending && report.id) {
      downloadMutation.mutate(report.id)
    }
  }

  return (
    <Card
      data-testid={`report-card-${report.id}`}
      className={cn(
        'relative h-[214px] w-[396px] overflow-hidden rounded-lg border border-[#F4F4F5] bg-white p-0 shadow-sm',
        className
      )}
      {...props}
    >
      {/* Content wrapper with proper padding from Figma: 16px top/bottom, 20px left/right */}
      <div className="flex h-full flex-col gap-2.5 p-4 pt-4 pr-5 pb-5 pl-5">
        {/* Card Header */}
        <CardHeader className="flex-row items-start justify-between gap-2 p-0">
          <div className="flex gap-2">
            {/* File icon - 32x32 from Figma */}
            <div className="flex h-8 w-8 items-center justify-center">
              <FileText className="h-4 w-4 text-[#A1A1AA]" strokeWidth={1.5} />
            </div>

            {/* Header content */}
            <div className="flex flex-col gap-2 pt-2">
              {/* Title with Figma typography: Inter 600, 14px, line-height 1.4 */}
              <h3 className="text-sm leading-[1.4] font-semibold text-[#3F3F46]">
                {report.id}
              </h3>

              {/* Badges row */}
              <div className="flex gap-2">
                <ReportStatusBadge status={report.status} />
                <Badge
                  variant="outline"
                  className="rounded-[10px] border border-[#F4F4F5] bg-white px-2 py-0.5 text-xs font-medium text-[#3F3F46]"
                >
                  {report.template?.outputFormat}
                </Badge>
              </div>
            </div>
          </div>

          {/* More actions button - positioned absolutely as in Figma */}
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                data-testid={`report-card-actions-${report.id}`}
                variant="ghost"
                size="sm"
                className="h-8 w-8 rounded-md p-0"
              >
                <MoreVertical
                  className="h-4 w-4 text-[#52525B]"
                  strokeWidth={1.5}
                />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              {isDownloadable && (
                <DropdownMenuItem
                  data-testid={`report-card-download-${report.id}`}
                  onClick={handleDownload}
                  disabled={downloadMutation.isPending}
                >
                  <Download className="mr-2 h-4 w-4" />
                  {intl.formatMessage({
                    id: 'reports.actions.download',
                    defaultMessage: 'Download Report'
                  })}
                </DropdownMenuItem>
              )}
              {!isDownloadable && (
                <DropdownMenuItem
                  data-testid={`report-card-download-disabled-${report.id}`}
                  disabled
                >
                  <Download className="mr-2 h-4 w-4" />
                  {intl.formatMessage({
                    id: 'reports.actions.downloadNotAvailable',
                    defaultMessage: 'Download Not Available'
                  })}
                </DropdownMenuItem>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        </CardHeader>

        {/* Card Body - fills remaining space */}
        <CardContent className="flex-1 p-0">
          <div className="flex flex-col gap-2">
            {/* Template name with Figma typography: Inter 500, 12px */}
            <div className="text-xs font-medium text-[#71717A]">
              {report.template?.name ?? report.template?.fileName}
            </div>
          </div>
        </CardContent>

        {/* Card Footer with border-top from Figma */}
        <CardFooter className="border-t border-[#E4E4E7] p-0 pt-2">
          <div className="flex items-center gap-2 text-xs font-medium text-[#A1A1AA]">
            <Clock className="h-3 w-3" strokeWidth={1.5} />
            {report.createdAt
              ? dayjs(report.createdAt).format('L - LT')
              : 'Unknown date'}
          </div>
        </CardFooter>
      </div>

      {/* Processing loader overlay */}
      {report.status === 'Processing' && (
        <div className="absolute right-0 bottom-0 left-0 h-1 overflow-hidden bg-[#F4F4F5]">
          <div className="h-full animate-pulse bg-blue-500" />
        </div>
      )}

      {/* Downloading overlay */}
      <DownloadingOverlay isDownloading={downloadMutation.isPending} />
    </Card>
  )
}
