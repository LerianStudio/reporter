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
  UseQueryOptions
} from '@tanstack/react-query'
import { PaginationRequest } from '@/types/pagination-request'
import { getRuntimeEnv } from '@lerianstudio/console-layout'
import { TemplateQueryKeys } from '@/lib/utils'
import { useRetryInvalidation } from '@/hooks/use-retry-invalidation'

const basePath =
  getRuntimeEnv('NEXT_PUBLIC_PLUGIN_UI_BASE_PATH') ??
  process.env.NEXT_PUBLIC_PLUGIN_UI_BASE_PATH

type UseListTemplatesProps = {
  organizationId?: string
  filters?: TemplateFiltersDto & PaginationRequest
} & PaginationRequest &
  Omit<UseQueryOptions<PaginationDto<TemplateDto>>, 'queryKey' | 'queryFn'>

type UseCreateTemplateProps = UseMutationOptions<
  TemplateDto,
  Error,
  FormData
> & {
  organizationId: string
  onSuccess?: (data: TemplateDto) => void
}

type UseGetTemplateProps = {
  organizationId: string
  templateId: string
  enabled?: boolean
}

type UseUpdateTemplateProps = UseMutationOptions<
  TemplateDto,
  Error,
  FormData
> & {
  organizationId: string
  templateId: string
  onSuccess?: (data: TemplateDto) => void
}

type UseDeleteTemplateProps = UseMutationOptions<
  void,
  Error,
  { id: string }
> & {
  organizationId: string
  templateId: string
  onSuccess?: () => void
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

  // Use centralized query key generation for consistency
  const queryKey = TemplateQueryKeys.list(organizationId || '', filters)

  return useQuery<PaginationDto<TemplateDto>>({
    queryKey,
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
    queryKey: TemplateQueryKeys.detail(organizationId, templateId),
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
  const { simpleInvalidateWithRetry } = useRetryInvalidation()

  return useMutation<TemplateDto, Error, FormData>({
    mutationKey: TemplateQueryKeys.mutations.create(organizationId),
    mutationFn: postFormDataFetcher(
      `${basePath}/api/organizations/${organizationId}/templates`
    ),
    onSuccess: async (data) => {
      await simpleInvalidateWithRetry(
        [TemplateQueryKeys.allLists(organizationId)],
        () => onSuccess?.(data)
      )
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
  const { invalidateWithRetry } = useRetryInvalidation()

  return useMutation<TemplateDto, Error, FormData>({
    mutationKey: TemplateQueryKeys.mutations.update(organizationId, templateId),
    mutationFn: patchFormDataFetcher(
      `${basePath}/api/organizations/${organizationId}/templates/${templateId}`
    ),
    onSuccess: async (data) => {
      await invalidateWithRetry({
        queryKeys: [
          TemplateQueryKeys.allLists(organizationId),
          TemplateQueryKeys.detail(organizationId, templateId)
        ],
        onSuccess: () => onSuccess?.(data)
      })
    },
    ...options
  })
}

export const useDeleteTemplate = ({
  organizationId,
  onSuccess,
  ...options
}: UseDeleteTemplateProps) => {
  const { simpleInvalidateWithRetry } = useRetryInvalidation()

  return useMutation<void, Error, { id: string }>({
    mutationKey: TemplateQueryKeys.mutations.delete(organizationId),
    mutationFn: deleteFetcher(
      `${basePath}/api/organizations/${organizationId}/templates`
    ),
    onSuccess: async () => {
      await simpleInvalidateWithRetry(
        [TemplateQueryKeys.allLists(organizationId)],
        onSuccess,
        1 // Single retry for delete operations (less critical)
      )
    },
    ...options
  })
}
