import {
  ReportStatus,
  ReportFilters,
  AdvancedReportFilters
} from '@/core/domain/entities/report-entity'
import { MetadataEntity } from '@/core/domain/entities/metadata-entity'
import { TemplateDto } from './template-dto'
import { SearchParamDto } from './request-dto'

export type ReportSearchParamDto = SearchParamDto & {
  status?: ReportStatus
  search?: string
  templateId?: string
}

export type CreateReportDto = {
  templateId: string
  filters?: ReportFilters
  metadata?: MetadataEntity
}

export type CreateAdvancedReportDto = {
  templateId: string
  filters?: AdvancedReportFilters
  metadata?: MetadataEntity
}

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
