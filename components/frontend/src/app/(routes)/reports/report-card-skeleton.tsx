import React from 'react'
import { Card } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils'

type ReportCardSkeletonProps = {
  className?: string
}

export function ReportCardSkeleton({ className }: ReportCardSkeletonProps) {
  return (
    <Card
      className={cn(
        'h-[214px] w-[396px] overflow-hidden rounded-lg border border-[#F4F4F5] bg-white p-0 shadow-sm',
        className
      )}
    >
      <div className="flex h-full flex-col gap-2.5 p-4 pt-4 pr-5 pb-5 pl-5">
        <div className="flex items-start justify-between gap-2">
          <div className="flex gap-2">
            <Skeleton className="h-8 w-8 rounded-sm" />

            <div className="flex flex-col gap-2 pt-2">
              <Skeleton className="h-4 w-32" />

              <div className="flex gap-2">
                <Skeleton className="h-5 w-20 rounded-[10px]" />
                <Skeleton className="h-5 w-16 rounded-[10px]" />
              </div>
            </div>
          </div>

          <Skeleton className="h-8 w-8 rounded-md" />
        </div>

        <div className="flex-1">
          <div className="flex flex-col gap-2">
            <Skeleton className="h-3 w-48" />
          </div>
        </div>

        <div className="border-t border-[#E4E4E7] pt-2">
          <div className="flex items-center gap-2">
            <Skeleton className="h-3 w-3 rounded-full" />
            <Skeleton className="h-3 w-28" />
          </div>
        </div>
      </div>
    </Card>
  )
}

type ReportGridSkeletonProps = {
  count?: number
}

export function ReportGridSkeleton({ count = 6 }: ReportGridSkeletonProps) {
  return (
    <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
      {Array.from({ length: count }, (_, index) => (
        <ReportCardSkeleton key={index} className="w-full max-w-none" />
      ))}
    </div>
  )
}
