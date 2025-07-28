import {
  ReportStatus,
  ReportFilters
} from '@/core/domain/entities/report-entity'
import { MetadataEntity } from '@/core/domain/entities/metadata-entity'
import { TemplateDto } from './template-dto'

/**
 * DTO for creating a new report generation request
 */
export type CreateReportDto = {
  templateId: string
  organizationId: string
  filters?: ReportFilters
  metadata?: MetadataEntity
}

/**
 * DTO for report responses
 */
export type ReportDto = {
  id: string
  templateId: string
  organizationId: string
  status?: ReportStatus
  filters?: ReportFilters
  template?: TemplateDto
  metadata?: MetadataEntity
  error?: string
  createdAt: Date
  updatedAt: Date
  completedAt?: Date
}

/**
 * DTO for complex report filtering operations
 */
export type ReportFiltersDto = {
  date_range?: {
    start: string // ISO 8601 date string
    end: string // ISO 8601 date string
  }
  account_types?: string[]
  minimum_balance?: number
  maximum_balance?: number
  asset_codes?: string[]
  portfolio_ids?: string[]
  search?: string
}

/**
 * DTO for report generation request with template information
 */
export type ReportGenerationRequestDto = {
  templateId: string
  templateName?: string
  filters?: ReportFiltersDto
  organizationId: string
}
