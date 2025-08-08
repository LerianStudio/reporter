import { PaginationSearchEntity } from './pagination-entity'

// Output format types for template generation
export type OutputFormat = 'csv' | 'xml' | 'html' | 'txt'

export type TemplateSearchEntity = PaginationSearchEntity & {
  outputFormat?: OutputFormat
  name?: string
}

// Main template entity
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
