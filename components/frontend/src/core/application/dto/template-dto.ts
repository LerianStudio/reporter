import { OutputFormat } from '@/core/domain/entities/template-entity'
import { SearchParamDto } from './request-dto'

export type TemplateSearchParamDto = SearchParamDto & {
  outputFormat?: OutputFormat
  name?: string
}

/**
 * DTO for creating a new template
 */
export type CreateTemplateDto = {
  organizationId: string
  name: string
  outputFormat: OutputFormat
  templateFile: File
}

/**
 * DTO for updating an existing template
 */
export type UpdateTemplateDto = {
  name?: string
  outputFormat?: OutputFormat
  templateFile?: File
}

/**
 * DTO for template responses
 */
export type TemplateDto = {
  id: string
  organizationId: string
  name: string
  fileName: string
  outputFormat: OutputFormat
  createdAt: Date
  updatedAt: Date
}

export type TemplateFiltersDto = {
  name?: string
  outputFormat?: OutputFormat
}
