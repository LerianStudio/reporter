import { ReporterPaginationMapper } from './reporter-pagination-mapper'
import { PaginationEntity } from '@/core/domain/entities/pagination-entity'
import { ReporterPaginationDto } from '../dto/reporter-pagination-dto'

describe('ReporterPaginationMapper', () => {
  it('should map PaginationEntity to PaginationDto with default mapper', () => {
    const paginationEntity: PaginationEntity<number> = {
      items: [1, 2, 3],
      limit: 10,
      page: 1
    }

    const result: ReporterPaginationDto<number> =
      ReporterPaginationMapper.toResponseDto(paginationEntity, (item) => item)

    expect(result).toEqual({
      items: [1, 2, 3],
      limit: 10,
      page: 1
    })
  })

  it('should map PaginationEntity to PaginationDto with custom mapper', () => {
    const paginationEntity: PaginationEntity<number> = {
      items: [1, 2, 3],
      limit: 10,
      page: 1
    }

    const customMapper = (item: number) => `Item ${item}`
    const result: ReporterPaginationDto<string> =
      ReporterPaginationMapper.toResponseDto(paginationEntity, customMapper)

    expect(result).toEqual({
      items: ['Item 1', 'Item 2', 'Item 3'],
      limit: 10,
      page: 1
    })
  })

  it('should handle empty items in PaginationEntity', () => {
    const paginationEntity: PaginationEntity<number> = {
      items: [],
      limit: 10,
      page: 1
    }

    const result: ReporterPaginationDto<number> =
      ReporterPaginationMapper.toResponseDto(paginationEntity, (item) => item)

    expect(result).toEqual({
      items: [],
      limit: 10,
      page: 1
    })
  })

  it('should handle null items in PaginationEntity', () => {
    const paginationEntity: PaginationEntity<number> = {
      items: null as any,
      limit: 10,
      page: 1
    }

    const result: ReporterPaginationDto<number> =
      ReporterPaginationMapper.toResponseDto(paginationEntity, (item) => item)

    expect(result).toEqual({
      items: [],
      limit: 10,
      page: 1
    })
  })
})
