import { PaginationEntity } from '@/core/domain/entities/pagination-entity'
import { SmartPaginationDto } from '../dto/smart-pagination-dto'

export class SmartPaginationMapper {
  static toResponseDto<T, R = T>(
    dto: SmartPaginationDto<T>,
    mapper: (item: T) => R
  ): PaginationEntity<R> {
    const items = dto.items ? dto.items.map(mapper) : []

    return {
      items: items,
      limit: dto.limit,
      page: dto.page
    }
  }
}
