import { SmartPaginationMapper } from './smart-pagination-mapper'
import { PaginationEntity } from '@/core/domain/entities/pagination-entity'
import { SmartPaginationDto } from '../dto/smart-pagination-dto'

describe('SmartPaginationMapper', () => {
  it('should map PaginationEntity to PaginationDto with default mapper', () => {
    const paginationEntity: PaginationEntity<number> = {
      items: [1, 2, 3],
      limit: 10,
      page: 1
    }

    const result: SmartPaginationDto<number> =
      SmartPaginationMapper.toResponseDto(paginationEntity, (item) => item)

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
    const result: SmartPaginationDto<string> =
      SmartPaginationMapper.toResponseDto(paginationEntity, customMapper)

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

    const result: SmartPaginationDto<number> =
      SmartPaginationMapper.toResponseDto(paginationEntity, (item) => item)

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

    const result: SmartPaginationDto<number> =
      SmartPaginationMapper.toResponseDto(paginationEntity, (item) => item)

    expect(result).toEqual({
      items: [],
      limit: 10,
      page: 1
    })
  })
})
