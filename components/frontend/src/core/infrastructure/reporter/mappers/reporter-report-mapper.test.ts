import { ReporterReportMapper } from './reporter-report-mapper'
import { ReportEntity } from '@/core/domain/entities/report-entity'

describe('ReporterReportMapper', () => {
  describe('toCreateDto', () => {
    it('should handle AdvancedReportFilters format directly', () => {
      const entity: ReportEntity = {
        templateId: '01992b60-c374-7656-9ed4-a60a36b3b1cd',
        organizationId: '019905b6-6793-7a4e-b568-291924bbf3ff',
        filters: {
          midaz_transaction: {
            balance: {
              available: {
                gt: ['1000.00']
              }
            }
          }
        } as any
      }

      const result = ReporterReportMapper.toCreateDto(entity)

      expect(result).toEqual({
        templateId: '01992b60-c374-7656-9ed4-a60a36b3b1cd',
        filters: {
          midaz_transaction: {
            balance: {
              available: {
                gt: ['1000.00']
              }
            }
          }
        },
        metadata: undefined
      })
    })

    it('should handle old fields array format', () => {
      const entity: ReportEntity = {
        templateId: '01992b60-c374-7656-9ed4-a60a36b3b1cd',
        organizationId: '019905b6-6793-7a4e-b568-291924bbf3ff',
        filters: {
          fields: [
            {
              database: 'midaz_transaction',
              table: 'balance',
              field: 'available',
              operator: 'gt',
              values: ['1000.00']
            }
          ]
        } as any
      }

      const result = ReporterReportMapper.toCreateDto(entity)

      expect(result).toEqual({
        templateId: '01992b60-c374-7656-9ed4-a60a36b3b1cd',
        filters: {
          midaz_transaction: {
            balance: {
              available: {
                gt: ['1000.00']
              }
            }
          }
        },
        metadata: undefined
      })
    })

    it('should return empty filters when no filters provided', () => {
      const entity: ReportEntity = {
        templateId: '01992b60-c374-7656-9ed4-a60a36b3b1cd',
        organizationId: '019905b6-6793-7a4e-b568-291924bbf3ff'
      }

      const result = ReporterReportMapper.toCreateDto(entity)

      expect(result).toEqual({
        templateId: '01992b60-c374-7656-9ed4-a60a36b3b1cd',
        filters: {},
        metadata: undefined
      })
    })
  })
})
