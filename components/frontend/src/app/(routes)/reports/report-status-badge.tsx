'use client'

import React from 'react'
import { useIntl, defineMessages } from 'react-intl'
import { cva, type VariantProps } from 'class-variance-authority'
import { Check, Clock, AlertTriangle } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { ReportDto } from '@/core/application/dto/report-dto'
import { cn } from '@/lib/utils'

const statusLabels = defineMessages({
  Finished: {
    id: 'reports.status.finished',
    defaultMessage: 'Finished'
  },
  Processing: {
    id: 'reports.status.processing',
    defaultMessage: 'Processing'
  },
  Failed: {
    id: 'reports.status.failed',
    defaultMessage: 'Failed'
  }
})

const reportStatusBadgeVariants = cva(
  'gap-1 rounded-[10px] border-none px-2 py-0.5 text-xs font-medium',
  {
    variants: {
      status: {
        Finished: 'bg-[#F0FDF4] text-[#166534]',
        Processing: 'bg-yellow-50 text-yellow-700',
        Failed: 'bg-red-50 text-red-700'
      }
    },
    defaultVariants: {
      status: 'Processing'
    }
  }
)

export type ReportStatusBadgeProps = VariantProps<
  typeof reportStatusBadgeVariants
> & {
  status: ReportDto['status']
  className?: string
}

export function ReportStatusBadge({
  status,
  className,
  ...props
}: ReportStatusBadgeProps) {
  const intl = useIntl()

  if (!status) {
    return null
  }

  const getIcon = () => {
    switch (status) {
      case 'Finished':
        return <Check className="h-3 w-3" />
      case 'Processing':
        return <Clock className="h-3 w-3" />
      case 'Failed':
        return <AlertTriangle className="h-3 w-3" />
      default:
        return null
    }
  }

  const getStatusLabel = () => {
    const validStatus = Object.keys(statusLabels).includes(status)
      ? (status as keyof typeof statusLabels)
      : 'Processing'

    return intl.formatMessage(statusLabels[validStatus])
  }

  return (
    <Badge
      variant="secondary"
      className={cn(reportStatusBadgeVariants({ status }), className)}
      {...props}
    >
      {getIcon()}
      {getStatusLabel()}
    </Badge>
  )
}
