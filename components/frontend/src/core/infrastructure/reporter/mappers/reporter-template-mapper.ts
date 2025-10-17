import { TemplateEntity } from '@/core/domain/entities/template-entity'
import { PaginationEntity } from '@/core/domain/entities/pagination-entity'
import {
  ReporterTemplateDto,
  ReporterCreateTemplateDto,
  ReporterUpdateTemplateDto
} from '../dto/reporter-template-dto'
import { ReporterPaginationDto } from '../dto/reporter-pagination-dto'
import { ReporterPaginationMapper } from './reporter-pagination-mapper'

/**
 * Mapper for converting between Reporter API DTOs and domain entities
 */
export class ReporterTemplateMapper {
  /**
   * Convert Reporter API DTO to domain entity
   */
  static toEntity(dto: ReporterTemplateDto): TemplateEntity {
    return {
      id: dto.id,
      organizationId: '', // Will be set from context
      name: dto.description || '',
      fileName: dto.fileName,
      outputFormat: dto.outputFormat,
      createdAt: dto.createdAt,
      updatedAt: dto.updatedAt
    }
  }

  /**
   * Convert template entity to Reporter API create DTO
   */
  static toCreateDto(entity: TemplateEntity): ReporterCreateTemplateDto {
    return {
      description: entity.name,
      outputFormat: entity.outputFormat,
      template: entity.templateFile as File
    }
  }

  /**
   * Convert partial template entity to Reporter API update DTO
   */
  static toUpdateDto(entity: Partial<TemplateEntity>): ReporterUpdateTemplateDto {
    return {
      description: entity.name,
      outputFormat: entity.outputFormat,
      template: entity.templateFile as File
    }
  }

  /**
   * Convert Reporter API pagination response to domain pagination entity
   */
  static toPaginationEntity(
    dto: ReporterPaginationDto<ReporterTemplateDto>
  ): PaginationEntity<TemplateEntity> {
    return ReporterPaginationMapper.toResponseDto(dto, this.toEntity)
  }
}
