import { OutputFormat } from '@/core/domain/entities/template-entity'

/**
 * DTO representing a template response from the Smart Templates API
 */
export type SmartTemplateDto = {
  id: string
  fileName: string
  description?: string
  outputFormat: OutputFormat
  createdAt: Date
  updatedAt?: Date
  deletedAt?: Date
}

/**
 * DTO for creating a new template via Smart Templates API
 */
export type SmartCreateTemplateDto = {
  description?: string
  outputFormat: OutputFormat
  template: File
}

/**
 * DTO for updating a template via Smart Templates API
 */
export type SmartUpdateTemplateDto = {
  description?: string
  outputFormat?: OutputFormat
  template?: File
}
