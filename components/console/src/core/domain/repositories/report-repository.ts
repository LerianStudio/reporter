import {
  ReportEntity,
  ReportStatus,
  ReportFilters
} from '../entities/report-entity'
import { PaginationEntity } from '../entities/pagination-entity'

/**
 * Extended filter parameters for querying reports
 */
export type ReportQueryFilters = ReportFilters & {
  /** Filter by report status */
  status?: ReportStatus
  /** Filter by template ID */
  templateId?: string
}

/**
 * Parameters for fetching reports with pagination and filtering
 */
export type FetchReportsParams = {
  /** Organization ID for multi-tenant isolation */
  organizationId: string
  /** Number of items per page */
  limit: number
  /** Page number (1-based) */
  page: number
  /** Optional filters */
  filters?: ReportQueryFilters
  /** Sort field */
  sortBy?: 'createdAt' | 'updatedAt' | 'completedAt' | 'status'
  /** Sort direction */
  sortOrder?: 'asc' | 'desc'
}

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
    params: FetchReportsParams
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
   * Fetch reports by status for an organization
   * @param organizationId Organization ID for access control
   * @param status Report status to filter by
   * @param limit Maximum number of results
   * @param page Page number for pagination
   * @returns Promise resolving to paginated list of reports with the specified status
   */
  abstract fetchByStatus(
    organizationId: string,
    status: ReportStatus,
    limit: number,
    page: number
  ): Promise<PaginationEntity<ReportEntity>>

  /**
   * Fetch reports by template ID for an organization
   * @param organizationId Organization ID for access control
   * @param templateId Template ID to filter by
   * @param limit Maximum number of results
   * @param page Page number for pagination
   * @returns Promise resolving to paginated list of reports for the specified template
   */
  abstract fetchByTemplate(
    organizationId: string,
    templateId: string,
    limit: number,
    page: number
  ): Promise<PaginationEntity<ReportEntity>>

  /**
   * Get download URL for a completed report
   * @param id Report ID
   * @param organizationId Organization ID for access control
   * @returns Promise resolving to the download URL
   * @throws Error if report not found, doesn't belong to organization, or is not downloadable
   */
  abstract getDownloadUrl(id: string, organizationId: string): Promise<string>

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

  /**
   * Count total reports for an organization
   * @param organizationId Organization ID
   * @param filters Optional filters to apply
   * @returns Promise resolving to total count
   */
  abstract countByOrganization(
    organizationId: string,
    filters?: ReportQueryFilters
  ): Promise<number>

  /**
   * Count reports by status for an organization
   * @param organizationId Organization ID
   * @param status Report status to count
   * @returns Promise resolving to count of reports with the specified status
   */
  abstract countByStatus(
    organizationId: string,
    status: ReportStatus
  ): Promise<number>

  /**
   * Search reports by text across filters and error messages
   * @param organizationId Organization ID for access control
   * @param searchText Text to search for
   * @param limit Maximum number of results
   * @param page Page number for pagination
   * @returns Promise resolving to paginated search results
   */
  abstract search(
    organizationId: string,
    searchText: string,
    limit: number,
    page: number
  ): Promise<PaginationEntity<ReportEntity>>

  /**
   * Fetch processing reports that may need status updates
   * @param organizationId Organization ID for access control
   * @param olderThan Optional timestamp to filter reports older than this time
   * @returns Promise resolving to list of processing reports
   */
  abstract fetchProcessingReports(
    organizationId: string,
    olderThan?: Date
  ): Promise<ReportEntity[]>
}
