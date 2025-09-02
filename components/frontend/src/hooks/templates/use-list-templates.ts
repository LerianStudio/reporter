import { TemplateFiltersDto } from '@/core/application/dto/template-dto'
import { useListTemplates as useBaseListTemplates } from '@/client/templates'

type UseListTemplatesWithDateProps = {
  organizationId?: string
  filters?: TemplateFiltersDto
  [key: string]: any
}

export const useListTemplates = ({
  organizationId,
  filters = {},
  ...options
}: UseListTemplatesWithDateProps) => {
  return useBaseListTemplates({
    organizationId,
    filters,
    staleTime: 30 * 1000,
    ...options
  })
}
