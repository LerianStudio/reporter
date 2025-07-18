import { TemplateEntity, TemplateFilters } from '../entities/template-entity'
import { PaginationEntity } from '../entities/pagination-entity'

/**
 * Parameters for fetching templates with pagination and filtering
 */
export interface FetchTemplatesParams {
  /** Organization ID for multi-tenant isolation */
  organizationId: string
  /** Number of items per page */
  limit: number
  /** Page number (1-based) */
  page: number
  /** Optional filters */
  filters?: TemplateFilters
}

/**
 * Template Repository Interface
 *
 * Defines the contract for template data access operations.
 * Follows Clean Architecture principles with domain-driven design.
 * Supports multi-tenant organization-based isolation.
 */
export abstract class TemplateRepository {
  /**
   * Create a new template
   * @param template Template entity to create
   * @returns Promise resolving to the created template with generated ID and timestamps
   * @throws Error if template with same fileName already exists in organization
   */
  abstract create(template: TemplateEntity): Promise<TemplateEntity>

  /**
   * Fetch all templates for an organization with pagination and filtering
   * @param params Fetch parameters including organizationId, pagination, and filters
   * @returns Promise resolving to paginated list of templates
   */
  abstract fetchAll(
    params: FetchTemplatesParams
  ): Promise<PaginationEntity<TemplateEntity>>

  /**
   * Fetch a template by ID with organization validation
   * @param id Template ID
   * @param organizationId Organization ID for access control
   * @returns Promise resolving to the template entity
   * @throws Error if template not found or doesn't belong to organization
   */
  abstract fetchById(
    id: string,
    organizationId: string
  ): Promise<TemplateEntity>

  /**
   * Update an existing template
   * @param id Template ID
   * @param organizationId Organization ID for access control
   * @param template Partial template data to update
   * @returns Promise resolving to the updated template
   * @throws Error if template not found, doesn't belong to organization, or fileName conflict
   */
  abstract update(
    id: string,
    organizationId: string,
    template: Partial<TemplateEntity>
  ): Promise<TemplateEntity>

  /**
   * Delete a template (soft delete)
   * @param id Template ID
   * @param organizationId Organization ID for access control
   * @returns Promise that resolves when deletion is complete
   * @throws Error if template not found or doesn't belong to organization
   */
  abstract delete(id: string, organizationId: string): Promise<void>

  /**
   * Count total templates for an organization
   * @param organizationId Organization ID
   * @param filters Optional filters to apply
   * @returns Promise resolving to total count
   */
  abstract countByOrganization(
    organizationId: string,
    filters?: TemplateFilters
  ): Promise<number>

  /**
   * Search templates by text across fileName and description
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
  ): Promise<PaginationEntity<TemplateEntity>>
}
