import {
  ReportStatus,
  ReportFilters
} from '@/core/domain/entities/report-entity'
import { MetadataEntity } from '@/core/domain/entities/metadata-entity'
import { TemplateDto } from './template-dto'
import { SearchParamDto } from './request-dto'

export type ReportSearchParamDto = SearchParamDto & {
  status?: ReportStatus
  search?: string
  templateId?: string
}

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
