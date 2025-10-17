import {
  ReportDto,
  CreateReportDto,
  CreateAdvancedReportDto,
  ReportFiltersDto
} from '@/core/application/dto/report-dto'
import { PaginationDto } from '@/core/application/dto/pagination-dto'
import {
  deleteFetcher,
  getFetcher,
  getPaginatedFetcher,
  postFetcher,
  downloadFetcher
} from '@/lib/fetcher'
import {
  useMutation,
  UseMutationOptions,
  useQuery,
  useQueryClient
} from '@tanstack/react-query'
import { getRuntimeEnv } from '@lerianstudio/console-layout'

const basePath =
  getRuntimeEnv('NEXT_PUBLIC_REPORTER_UI_BASE_PATH') ??
  process.env.NEXT_PUBLIC_REPORTER_UI_BASE_PATH

type UseListReportsProps = {
  organizationId?: string
  filters?: ReportFiltersDto & { limit?: number; page?: number }
  enabled?: boolean
}

type UseCreateReportProps = UseMutationOptions<any, any, any> & {
  organizationId: string
  onSuccess?: (...args: any[]) => void
}

type UseGetReportProps = {
  organizationId: string
  reportId: string
  enabled?: boolean
}

type UseDeleteReportProps = UseMutationOptions & {
  organizationId: string
  reportId: string
  onSuccess?: (...args: any[]) => void
}

type UseDownloadReportProps = UseMutationOptions & {
  organizationId: string
  reportId: string
  onSuccess?: (url: string) => void
}

export const useListReports = ({
  organizationId,
  filters,
  enabled = true,
  ...options
}: UseListReportsProps = {}) => {
  const queryParams = {
    limit: filters?.limit || 10,
    page: filters?.page || 1,
    ...(filters?.search && { search: filters.search }),
    ...(filters?.status && { status: filters.status }),
    ...(filters?.templateId && { templateId: filters.templateId }),
    ...(filters?.createdAt && { createdAt: filters.createdAt })
  }

  return useQuery<PaginationDto<ReportDto>>({
    queryKey: ['reports', organizationId, queryParams],
    queryFn: getPaginatedFetcher(
      `${basePath}/api/organizations/${organizationId}/reports`,
      queryParams
    ),
    enabled,
    ...options
  })
}

export const useGetReport = ({
  organizationId,
  reportId,
  enabled = true,
  ...options
}: UseGetReportProps) => {
  return useQuery<ReportDto>({
    queryKey: ['reports', reportId],
    queryFn: getFetcher(
      `${basePath}/api/organizations/${organizationId}/reports/${reportId}`
    ),
    enabled: enabled && !!reportId,
    ...options
  })
}

export const useCreateReport = ({
  organizationId,
  onSuccess,
  ...options
}: UseCreateReportProps) => {
  const queryClient = useQueryClient()

  return useMutation<any, any, CreateAdvancedReportDto>({
    mutationKey: ['reports'],
    mutationFn: postFetcher(
      `${basePath}/api/organizations/${organizationId}/reports`
    ),
    onSuccess: async (...args) => {
      await queryClient.invalidateQueries({
        queryKey: ['reports']
      })
      onSuccess?.(...args)
    },
    ...options
  })
}

export const useDeleteReport = ({
  organizationId,
  reportId,
  onSuccess,
  ...options
}: UseDeleteReportProps) => {
  const queryClient = useQueryClient()

  return useMutation<any, any, any>({
    mutationKey: ['reports', reportId],
    mutationFn: deleteFetcher(
      `${basePath}/api/organizations/${organizationId}/reports/${reportId}`
    ),
    onSuccess: async (...args) => {
      await queryClient.invalidateQueries({
        queryKey: ['reports']
      })
      onSuccess?.(...args)
    },
    ...options
  })
}

export const useDownloadReport = ({
  organizationId,
  reportId,
  onSuccess,
  ...options
}: UseDownloadReportProps) => {
  return useMutation<any, any, any>({
    mutationKey: ['reports', reportId, 'download'],
    mutationFn: downloadFetcher(
      `${basePath}/api/organizations/${organizationId}/reports/${reportId}/download`
    ),
    onSuccess: (...args) => {
      onSuccess?.(...args)
    },
    ...options
  })
}

export const useGetReportDownloadInfo = ({
  organizationId,
  reportId,
  enabled = true,
  ...options
}: UseGetReportProps) => {
  return useQuery<{ downloadUrl: string; fileName: string; fileSize: number }>({
    queryKey: ['reports', reportId, 'download-info'],
    queryFn: getFetcher(
      `${basePath}/api/organizations/${organizationId}/reports/${reportId}/download-info`
    ),
    enabled: enabled && !!reportId,
    ...options
  })
}
