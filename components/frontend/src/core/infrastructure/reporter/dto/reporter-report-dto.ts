import { ReportStatus } from '@/core/domain/entities/report-entity'
import { ReporterMetadataDto } from './reporter-metadata-dto'

/**
 * DTO representing a report response from the Reporter API
 */
export type ReporterReportDto = {
  id: string
  templateId: string
  status: ReportStatus
  filters?: Record<string, any>
  metadata?: ReporterMetadataDto
  error?: string
  createdAt: Date
  updatedAt?: Date
  completedAt?: Date
}

/**
 * DTO for creating a new report generation request via Reporter API
 * Matches the External API format: database -> table -> field -> values[]
 */
export type ReporterCreateReportDto = {
  templateId: string
  filters: Record<string, Record<string, Record<string, string[]>>>
  metadata?: ReporterMetadataDto
}
