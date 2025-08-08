import { TemplateEntity } from '@/core/domain/entities/template-entity'
import { PaginationEntity } from '@/core/domain/entities/pagination-entity'
import {
  SmartTemplateDto,
  SmartCreateTemplateDto,
  SmartUpdateTemplateDto
} from '../dto/smart-template-dto'
import { SmartPaginationDto } from '../dto/smart-pagination-dto'
import { SmartPaginationMapper } from './smart-pagination-mapper'

/**
 * Mapper for converting between Smart Templates API DTOs and domain entities
 */
export class SmartTemplateMapper {
  /**
   * Convert Smart Template API DTO to domain entity
   */
  static toEntity(dto: SmartTemplateDto): TemplateEntity {
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
   * Convert template entity to Smart Templates API create DTO
   */
  static toCreateDto(entity: TemplateEntity): SmartCreateTemplateDto {
    return {
      description: entity.name,
      outputFormat: entity.outputFormat,
      template: entity.templateFile as File
    }
  }

  /**
   * Convert partial template entity to Smart Templates API update DTO
   */
  static toUpdateDto(entity: Partial<TemplateEntity>): SmartUpdateTemplateDto {
    return {
      description: entity.name,
      outputFormat: entity.outputFormat,
      template: entity.templateFile as File
    }
  }

  /**
   * Convert Smart Templates API pagination response to domain pagination entity
   */
  static toPaginationEntity(
    dto: SmartPaginationDto<SmartTemplateDto>
  ): PaginationEntity<TemplateEntity> {
    return SmartPaginationMapper.toResponseDto(dto, this.toEntity)
  }
}
