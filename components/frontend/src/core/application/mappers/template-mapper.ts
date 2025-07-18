import { TemplateEntity } from '@/core/domain/entities/template-entity'
import { TemplateDto } from '@/core/application/dto/template-dto'
import { PaginationEntity } from '@/core/domain/entities/pagination-entity'
import { PaginationMapper } from './pagination-mapper'

/**
 * Mapper class for converting between Template entities and DTOs
 */
export class TemplateMapper {
  static toEntity(dto: TemplateDto): TemplateEntity {
    return {
      id: dto.id,
      organizationId: dto.organizationId,
      name: dto.name,
      outputFormat: dto.outputFormat
    }
  }

  /**
   * Convert a template entity to a response DTO
   */
  static toResponseDto(entity: TemplateEntity): TemplateDto {
    return {
      id: entity.id!,
      organizationId: entity.organizationId,
      fileName: entity.fileName!,
      name: entity.name,
      outputFormat: entity.outputFormat,
      createdAt: entity.createdAt!,
      updatedAt: entity.updatedAt!
    }
  }

  static toPaginationResponseDto(
    result: PaginationEntity<TemplateEntity>
  ): PaginationEntity<TemplateDto> {
    return PaginationMapper.toResponseDto(result, TemplateMapper.toResponseDto)
  }
}
