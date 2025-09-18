import { OutputFormat } from '@/core/domain/entities/template-entity'
import { SearchParamDto } from './request-dto'

export type TemplateSearchParamDto = SearchParamDto & {
  outputFormat?: OutputFormat
  name?: string
}

export type CreateTemplateDto = {
  organizationId: string
  name: string
  outputFormat: OutputFormat
  templateFile: File
}

export type UpdateTemplateDto = {
  name?: string
  outputFormat?: OutputFormat
  templateFile?: File
}

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
  createdAt?: string
}
