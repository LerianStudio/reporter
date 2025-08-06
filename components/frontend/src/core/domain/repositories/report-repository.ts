import {
  ReportEntity,
  ReportStatus,
  ReportFilters,
  ReportSearchEntity
} from '../entities/report-entity'
import { PaginationEntity } from '../entities/pagination-entity'

/**
 * File download response containing file content and metadata
 */
export type DownloadFileResponse = {
  /** File content as text */
  content: string
  /** Original filename */
  fileName: string
  /** MIME type of the file */
  contentType: string
}

/**
 * Report Repository Interface
 *
 * Defines the contract for report data access operations.
 * Follows Clean Architecture principles with domain-driven design.
 * Supports multi-tenant organization-based isolation and report lifecycle management.
 *
 * Note: The external Report API does not support update operations, so this interface
 * only provides read and create operations along with download functionality.
 */
export abstract class ReportRepository {
  /**
   * Create a new report generation request
   * @param report Report entity to create
   * @returns Promise resolving to the created report with generated ID and timestamps
   * @throws Error if template doesn't exist or doesn't belong to organization
   */
  abstract create(report: ReportEntity): Promise<ReportEntity>

  /**
   * Fetch all reports for an organization with pagination and filtering
   * @param params Fetch parameters including organizationId, pagination, and filters
   * @returns Promise resolving to paginated list of reports
   */
  abstract fetchAll(
    organizationId: string,
    query: ReportSearchEntity
  ): Promise<PaginationEntity<ReportEntity>>

  /**
   * Fetch a report by ID with organization validation
   * @param id Report ID
   * @param organizationId Organization ID for access control
   * @returns Promise resolving to the report entity
   * @throws Error if report not found or doesn't belong to organization
   */
  abstract fetchById(id: string, organizationId: string): Promise<ReportEntity>

  /**
   * Download file content directly from the external API
   * @param id Report ID
   * @param organizationId Organization ID for access control
   * @returns Promise resolving to file content and metadata
   * @throws Error if report not found, doesn't belong to organization, or is not downloadable
   */
  abstract downloadFile(
    id: string,
    organizationId: string
  ): Promise<DownloadFileResponse>
}
