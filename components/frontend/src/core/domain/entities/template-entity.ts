import { PaginationSearchEntity } from './pagination-entity'

export type OutputFormat = 'csv' | 'xml' | 'html' | 'txt'

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
