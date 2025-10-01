import { PaginationSearchEntity } from './pagination-entity'

export const OUTPUT_FORMATS = ['csv', 'xml', 'html', 'txt', 'pdf'] as const
export type OutputFormat = (typeof OUTPUT_FORMATS)[number]

export type TemplateSearchEntity = PaginationSearchEntity & {
  outputFormat?: OutputFormat
  name?: string
  createdAt?: Date
}

export type TemplateEntity = {
  id?: string
  organizationId: string
  name: string
  fileName?: string
  outputFormat: OutputFormat
  templateFile?: File
  createdAt?: Date
  updatedAt?: Date
}

export type TemplateFilters = {
  name?: string
  outputFormat?: OutputFormat
}
