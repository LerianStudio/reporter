import { PaginationEntity } from '@/core/domain/entities/pagination-entity'
import { ReporterPaginationDto } from '../dto/reporter-pagination-dto'

export class ReporterPaginationMapper {
  static toResponseDto<T, R = T>(
    dto: ReporterPaginationDto<T>,
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
