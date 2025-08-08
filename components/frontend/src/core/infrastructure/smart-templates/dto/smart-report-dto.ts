import { ReportStatus } from '@/core/domain/entities/report-entity'
import { SmartMetadataDto } from './smart-metadata-dto'

/**
 * DTO representing a report response from the Smart Templates API
 */
export type SmartReportDto = {
  id: string
  templateId: string
  status: ReportStatus
  filters?: Record<string, any>
  metadata?: SmartMetadataDto
  error?: string
  createdAt: Date
  updatedAt?: Date
  completedAt?: Date
}

/**
 * DTO for creating a new report generation request via Smart Templates API
 * Matches the External API format: database -> table -> field -> values[]
 */
export type SmartCreateReportDto = {
  templateId: string
  filters: Record<string, Record<string, Record<string, string[]>>>
  metadata?: SmartMetadataDto
}
