import { OutputFormat } from '@/core/domain/entities/template-entity'

/**
 * DTO representing a template response from the Reporter API
 */
export type ReporterTemplateDto = {
  id: string
  fileName: string
  description?: string
  outputFormat: OutputFormat
  createdAt: Date
  updatedAt?: Date
  deletedAt?: Date
}

/**
 * DTO for creating a new template via Reporter API
 */
export type ReporterCreateTemplateDto = {
  description?: string
  outputFormat: OutputFormat
  template: File
}

/**
 * DTO for updating a template via Reporter API
 */
export type ReporterUpdateTemplateDto = {
  description?: string
  outputFormat?: OutputFormat
  template?: File
}
