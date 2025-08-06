import {
  TemplateEntity,
  TemplateSearchEntity
} from '../entities/template-entity'
import { PaginationEntity } from '../entities/pagination-entity'

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
    organizationId: string,
    query: TemplateSearchEntity
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
}
