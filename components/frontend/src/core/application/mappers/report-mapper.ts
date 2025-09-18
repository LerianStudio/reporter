import { ReportEntity } from '@/core/domain/entities/report-entity'
import { CreateReportDto, ReportDto } from '@/core/application/dto/report-dto'
import { PaginationEntity } from '@/core/domain/entities/pagination-entity'
import { PaginationMapper } from './pagination-mapper'
import { TemplateMapper } from './template-mapper'

export class ReportMapper {
  static toEntity(dto: CreateReportDto, organizationId: string): ReportEntity {
    return {
      templateId: dto.templateId,
      organizationId,
      filters: {
        fields: dto.fields
      },
      metadata: dto.metadata
    }
  }

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

  static toPaginationResponseDto(
    result: PaginationEntity<ReportEntity>
  ): PaginationEntity<ReportDto> {
    return PaginationMapper.toResponseDto(result, ReportMapper.toResponseDto)
  }
}
