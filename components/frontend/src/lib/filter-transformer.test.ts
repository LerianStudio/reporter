import {
  transformFiltersToApiFormat,
  parseFilterValues,
  transformToApiPayload
} from './filter-transformer'
import { FilterField } from '@/core/domain/entities/report-entity'

describe('filter-transformer', () => {
  describe('parseFilterValues', () => {
    it('should parse comma-separated values correctly', () => {
      const result = parseFilterValues('value1, value2, value3')
      expect(result).toEqual(['value1', 'value2', 'value3'])
    })

    it('should handle values with extra spaces', () => {
      const result = parseFilterValues(' value1 , value2,  value3  ')
      expect(result).toEqual(['value1', 'value2', 'value3'])
    })

    it('should filter out empty values', () => {
      const result = parseFilterValues('value1, , value3,   ')
      expect(result).toEqual(['value1', 'value3'])
    })
  })

  describe('transformFiltersToApiFormat', () => {
    it('should transform simple filter to nested structure', () => {
      const fields: FilterField[] = [
        {
          database: 'midaz_onboarding',
          table: 'account',
          field: 'id',
          operator: 'eq',
          values: ['01986d42-df4b-7233-b1b1-39eba87dd965']
        }
      ]

      const result = transformFiltersToApiFormat(fields)

      expect(result).toEqual({
        midaz_onboarding: {
          account: {
            id: {
              eq: ['01986d42-df4b-7233-b1b1-39eba87dd965']
            }
          }
        }
      })
    })

    it('should handle multiple filters on same field with different operators', () => {
      const fields: FilterField[] = [
        {
          database: 'midaz_onboarding',
          table: 'account',
          field: 'created_at',
          operator: 'gte',
          values: ['2025-08-02']
        },
        {
          database: 'midaz_onboarding',
          table: 'account',
          field: 'created_at',
          operator: 'lte',
          values: ['2025-08-03']
        }
      ]

      const result = transformFiltersToApiFormat(fields)

      expect(result).toEqual({
        midaz_onboarding: {
          account: {
            created_at: {
              gte: ['2025-08-02'],
              lte: ['2025-08-03']
            }
          }
        }
      })
    })

    it('should handle multiple databases, tables, and fields', () => {
      const fields: FilterField[] = [
        {
          database: 'midaz_onboarding',
          table: 'account',
          field: 'id',
          operator: 'eq',
          values: ['01986d42-df4b-7233-b1b1-39eba87dd965']
        },
        {
          database: 'midaz_onboarding',
          table: 'asset',
          field: 'code',
          operator: 'in',
          values: ['BRL', 'USD']
        }
      ]

      const result = transformFiltersToApiFormat(fields)

      expect(result).toEqual({
        midaz_onboarding: {
          account: {
            id: {
              eq: ['01986d42-df4b-7233-b1b1-39eba87dd965']
            }
          },
          asset: {
            code: {
              in: ['BRL', 'USD']
            }
          }
        }
      })
    })
  })

  describe('transformToApiPayload', () => {
    it('should create complete API payload', () => {
      const templateId = 'template-123'
      const fields: FilterField[] = [
        {
          database: 'midaz_onboarding',
          table: 'account',
          field: 'id',
          operator: 'eq',
          values: ['01986d42-df4b-7233-b1b1-39eba87dd965']
        }
      ]

      const result = transformToApiPayload(templateId, fields)

      expect(result).toEqual({
        templateId: 'template-123',
        filters: {
          midaz_onboarding: {
            account: {
              id: {
                eq: ['01986d42-df4b-7233-b1b1-39eba87dd965']
              }
            }
          }
        }
      })
    })

  })
})
