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
      organizationId: '', // Will be set from context
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
   * Transforms fields array to nested database -> table -> field -> values[] structure
   */
  static toCreateDto(entity: ReportEntity): SmartCreateReportDto {
    // Transform ledger_ids from filters to ledgerId array
    const ledgerIds = entity.filters?.ledger_ids || []

    // Build the nested filters structure from the fields array
    const nestedFilters: Record<
      string,
      Record<string, Record<string, string[]>>
    > = {}

    if (entity.filters?.fields && entity.filters.fields.length > 0) {
      // Iterate over each filter field and build the nested structure
      entity.filters.fields.forEach((filterField) => {
        const { database, table, field, values } = filterField

        // Initialize nested structure if it doesn't exist
        if (!nestedFilters[database]) {
          nestedFilters[database] = {}
        }
        if (!nestedFilters[database][table]) {
          nestedFilters[database][table] = {}
        }

        // Set the field values
        nestedFilters[database][table][field] = values
      })
    }

    return {
      templateId: entity.templateId,
      ledgerId: ledgerIds,
      filters: nestedFilters,
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
