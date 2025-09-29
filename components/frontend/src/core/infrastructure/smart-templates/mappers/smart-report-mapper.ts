import { ReportEntity } from '@/core/domain/entities/report-entity'
import { PaginationEntity } from '@/core/domain/entities/pagination-entity'
import { SmartReportDto, SmartCreateReportDto } from '../dto/smart-report-dto'
import { SmartPaginationDto } from '../dto/smart-pagination-dto'
import { SmartPaginationMapper } from './smart-pagination-mapper'

/**
 * Mapper for converting between Smart Templates API DTOs and domain entities
 */
export class SmartReportMapper {
  /**
   * Convert Smart Report API DTO to domain entity
   */
  static toEntity(dto: SmartReportDto): ReportEntity {
    return {
      id: dto.id,
      templateId: dto.templateId,
      organizationId: '',
      status: dto.status,
      filters: dto.filters,
      metadata: dto.metadata,
      createdAt: new Date(dto.createdAt),
      updatedAt: dto.updatedAt ? new Date(dto.updatedAt) : undefined,
      completedAt: dto.completedAt ? new Date(dto.completedAt) : undefined
    }
  }

  /**
   * Convert report entity to Smart Templates API create DTO
   * Handles both old fields array format and new nested AdvancedReportFilters structure
   */
  static toCreateDto(entity: ReportEntity): SmartCreateReportDto {
    let filters: any = {}

    if (entity.filters && typeof entity.filters === 'object') {
      const filterKeys = Object.keys(entity.filters)
      const isOldFormat = filterKeys.some((key) =>
        [
          'fields',
          'date_range',
          'account_types',
          'minimum_balance',
          'maximum_balance',
          'asset_codes',
          'portfolio_ids',
          'search'
        ].includes(key)
      )

      if (
        isOldFormat &&
        entity.filters.fields &&
        entity.filters.fields.length > 0
      ) {
        entity.filters.fields.forEach((filterField) => {
          const { database, table, field, operator, values } = filterField

          if (!filters[database]) {
            filters[database] = {}
          }
          if (!filters[database][table]) {
            filters[database][table] = {}
          }
          if (!filters[database][table][field]) {
            filters[database][table][field] = {}
          }

          filters[database][table][field][operator] = Array.isArray(values)
            ? values
            : [values]
        })
      } else if (!isOldFormat) {
        filters = entity.filters
      }
    }

    return {
      templateId: entity.templateId,
      filters: Object.keys(filters).length > 0 ? filters : {},
      metadata: entity.metadata || undefined
    }
  }

  /**
   * Convert Smart Templates API pagination response to domain pagination entity
   */
  static toPaginationEntity(
    dto: SmartPaginationDto<SmartReportDto>
  ): PaginationEntity<ReportEntity> {
    return SmartPaginationMapper.toResponseDto(dto, this.toEntity)
  }
}
