import {
  TemplateDto,
  TemplateFiltersDto
} from '@/core/application/dto/template-dto'
import { PaginationDto } from '@/core/application/dto/pagination-dto'
import {
  deleteFetcher,
  getFetcher,
  getPaginatedFetcher,
  postFormDataFetcher,
  patchFormDataFetcher
} from '@/lib/fetcher'
import {
  useMutation,
  UseMutationOptions,
  useQuery,
  useQueryClient
} from '@tanstack/react-query'
import { PaginationRequest } from '@/types/pagination-request'
import { getRuntimeEnv } from '@lerianstudio/console-layout'

const basePath = getRuntimeEnv('NEXT_PUBLIC_PLUGIN_UI_BASE_PATH')

type UseListTemplatesProps = {
  organizationId?: string
  filters?: TemplateFiltersDto & PaginationRequest
} & PaginationRequest

type UseCreateTemplateProps = UseMutationOptions<any, any, any> & {
  organizationId: string
  onSuccess?: (...args: any[]) => void
}

type UseGetTemplateProps = {
  organizationId: string
  templateId: string
  enabled?: boolean
}

type UseUpdateTemplateProps = UseMutationOptions & {
  organizationId: string
  templateId: string
  onSuccess?: (...args: any[]) => void
}

type UseDeleteTemplateProps = UseMutationOptions & {
  organizationId: string
  templateId: string
  onSuccess?: (...args: any[]) => void
}

export const useListTemplates = ({
  organizationId,
  filters,
  ...options
}: UseListTemplatesProps = {}) => {
  const queryParams = {
    limit: filters?.limit,
    page: filters?.page,
    ...(filters && { ...filters })
  }

  return useQuery<PaginationDto<TemplateDto>>({
    queryKey: ['templates', organizationId, Object.values(filters || {})],
    queryFn: getPaginatedFetcher(
      `${basePath}/api/organizations/${organizationId}/templates`,
      queryParams
    ),
    ...options
  })
}

export const useGetTemplate = ({
  organizationId,
  templateId,
  enabled = true,
  ...options
}: UseGetTemplateProps) => {
  return useQuery<TemplateDto>({
    queryKey: ['templates', templateId],
    queryFn: getFetcher(
      `${basePath}/api/organizations/${organizationId}/templates/${templateId}`
    ),
    enabled: enabled && !!templateId,
    ...options
  })
}

export const useCreateTemplate = ({
  organizationId,
  onSuccess,
  ...options
}: UseCreateTemplateProps) => {
  const queryClient = useQueryClient()

  return useMutation<any, any, any>({
    mutationKey: ['templates'],
    mutationFn: postFormDataFetcher(
      `${basePath}/api/organizations/${organizationId}/templates`
    ),
    onSuccess: async (...args) => {
      await queryClient.invalidateQueries({
        queryKey: ['templates']
      })
      onSuccess?.(...args)
    },
    ...options
  })
}

export const useUpdateTemplate = ({
  organizationId,
  templateId,
  onSuccess,
  ...options
}: UseUpdateTemplateProps) => {
  const queryClient = useQueryClient()

  return useMutation<any, any, any>({
    mutationKey: ['templates', templateId],
    mutationFn: patchFormDataFetcher(
      `${basePath}/api/organizations/${organizationId}/templates/${templateId}`
    ),
    onSuccess: async (...args) => {
      await queryClient.invalidateQueries({
        queryKey: ['templates']
      })
      await queryClient.invalidateQueries({
        queryKey: ['templates', templateId]
      })
      onSuccess?.(...args)
    },
    ...options
  })
}

export const useDeleteTemplate = ({
  organizationId,
  onSuccess,
  ...options
}: UseDeleteTemplateProps) => {
  const queryClient = useQueryClient()

  return useMutation<any, any, any>({
    mutationKey: ['templates', organizationId],
    mutationFn: deleteFetcher(
      `${basePath}/api/organizations/${organizationId}/templates`
    ),
    onSuccess: async (...args) => {
      await queryClient.invalidateQueries({
        queryKey: ['templates']
      })
      onSuccess?.(...args)
    },
    ...options
  })
}
