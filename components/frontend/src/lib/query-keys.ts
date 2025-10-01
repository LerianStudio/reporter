import { TemplateFiltersDto } from '@/core/application/dto/template-dto'
import { PaginationRequest } from '@/types/pagination-request'

type FilterValue = string | number | boolean | undefined | null

/**
 * Utility to create consistent query keys for templates
 */
export class TemplateQueryKeys {
  private static readonly baseKey = ['templates'] as const

  /**
   * Generate list query key with normalized filters
   */
  static list(
    organizationId: string,
    filters?: TemplateFiltersDto & PaginationRequest
  ): (string | Record<string, FilterValue>)[] {
    const baseListKey = [...this.baseKey, 'list', organizationId] as const

    if (!filters) {
      return [...baseListKey]
    }

    // Create normalized filters object with only defined values
    const normalizedFilters = this.normalizeFilters(filters)

    // Return key with filters only if there are any defined values
    return Object.keys(normalizedFilters).length > 0
      ? [...baseListKey, normalizedFilters]
      : [...baseListKey]
  }

  /**
   * Generate detail query key
   */
  static detail(organizationId: string, templateId: string): string[] {
    return [...this.baseKey, 'detail', organizationId, templateId]
  }

  /**
   * Generate mutation keys
   */
  static mutations = {
    create: (organizationId: string) => [
      ...this.baseKey,
      'create',
      organizationId
    ],
    update: (organizationId: string, templateId: string) => [
      ...this.baseKey,
      'update',
      organizationId,
      templateId
    ],
    delete: (organizationId: string) => [
      ...this.baseKey,
      'delete',
      organizationId
    ]
  }

  /**
   * Get all template-related query keys for an organization (for invalidation)
   */
  static all(organizationId: string): string[] {
    return [...this.baseKey, organizationId]
  }

  /**
   * Get all list query keys for an organization (for targeted invalidation)
   */
  static allLists(organizationId: string): string[] {
    return [...this.baseKey, 'list', organizationId]
  }

  /**
   * Normalize filters to ensure consistent query keys
   * Only includes defined, non-empty values
   */
  private static normalizeFilters(
    filters: Record<string, FilterValue>
  ): Record<string, FilterValue> {
    const normalized: Record<string, FilterValue> = {}

    Object.entries(filters).forEach(([key, value]) => {
      // Only include defined, non-null, non-empty values
      if (this.isValidFilterValue(value)) {
        normalized[key] = value
      }
    })

    return normalized
  }

  /**
   * Check if a filter value should be included in the query key
   */
  private static isValidFilterValue(
    value: FilterValue
  ): value is string | number | boolean {
    if (value === null || value === undefined) {
      return false
    }

    if (typeof value === 'string' && value.trim() === '') {
      return false
    }

    if (typeof value === 'number' && (!isFinite(value) || isNaN(value))) {
      return false
    }

    return true
  }
}
