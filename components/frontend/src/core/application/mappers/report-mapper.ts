import { ReportEntity } from '@/core/domain/entities/report-entity'
import { ReportDto } from '@/core/application/dto/report-dto'
import { PaginationEntity } from '@/core/domain/entities/pagination-entity'
import { PaginationMapper } from './pagination-mapper'
import { TemplateMapper } from './template-mapper'

/**
 * Mapper class for converting between Report entities and DTOs
 */
export class ReportMapper {
  /**
   * Convert a report DTO to an entity
   */
  static toEntity(dto: ReportDto): ReportEntity {
    return {
      id: dto.id,
      templateId: dto.templateId,
      organizationId: dto.organizationId,
      status: dto.status,
      filters: dto.filters,
      template: dto.template
        ? TemplateMapper.toEntity(dto.template)
        : undefined,
      metadata: dto.metadata,
      createdAt: dto.createdAt,
      updatedAt: dto.updatedAt,
      completedAt: dto.completedAt
    }
  }

  /**
   * Convert a report entity to a response DTO
   */
  static toResponseDto(entity: ReportEntity): ReportDto {
    return {
      id: entity.id!,
      templateId: entity.templateId,
      organizationId: entity.organizationId,
      status: entity.status,
      filters: entity.filters,
      template: entity.template
        ? TemplateMapper.toResponseDto(entity.template)
        : undefined,
      metadata: entity.metadata,
      createdAt: entity.createdAt!,
      updatedAt: entity.updatedAt!,
      completedAt: entity.completedAt
    }
  }

  /**
   * Convert paginated report entities to paginated DTOs
   */
  static toPaginationResponseDto(
    result: PaginationEntity<ReportEntity>
  ): PaginationEntity<ReportDto> {
    return PaginationMapper.toResponseDto(result, ReportMapper.toResponseDto)
  }
}
