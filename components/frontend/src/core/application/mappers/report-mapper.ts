import {
  ReportEntity,
  ReportFilters
} from '@/core/domain/entities/report-entity'
import {
  ReportDto,
  ReportFiltersDto,
  CreateReportDto
} from '@/core/application/dto/report-dto'
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
   * Convert a CreateReportDto to an entity for creation
   */
  static fromCreateDto(dto: CreateReportDto): ReportEntity {
    return {
      templateId: dto.templateId,
      organizationId: dto.organizationId,
      filters: dto.filters,
      metadata: dto.metadata,
      createdAt: new Date()
    }
  }

  /**
   * Convert ReportFilters to ReportFiltersDto
   */
  static filtersToDto(filters: ReportFilters): ReportFiltersDto {
    return {
      ledger_ids: filters.ledger_ids,
      date_range: filters.date_range,
      account_types: filters.account_types,
      minimum_balance: filters.minimum_balance,
      maximum_balance: filters.maximum_balance,
      asset_codes: filters.asset_codes,
      portfolio_ids: filters.portfolio_ids,
      search: filters.search
    }
  }

  /**
   * Convert ReportFiltersDto to ReportFilters
   */
  static filtersFromDto(dto: ReportFiltersDto): ReportFilters {
    return {
      ledger_ids: dto.ledger_ids,
      date_range: dto.date_range,
      account_types: dto.account_types,
      minimum_balance: dto.minimum_balance,
      maximum_balance: dto.maximum_balance,
      asset_codes: dto.asset_codes,
      portfolio_ids: dto.portfolio_ids,
      search: dto.search
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
