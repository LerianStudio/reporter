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

    describe('comma-separated string parsing', () => {
      it('should split comma-separated string for "in" operator', () => {
        const entity: ReportEntity = {
          templateId: 'test-template-id',
          organizationId: 'test-org-id',
          filters: {
            fields: [
              {
                database: 'midaz_transaction',
                table: 'account',
                field: 'status',
                operator: 'in',
                values: 'active,pending,completed'
              }
            ]
          } as any
        }

        const result = ReporterReportMapper.toCreateDto(entity)

        expect(result.filters.midaz_transaction.account.status.in).toEqual([
          'active',
          'pending',
          'completed'
        ])
      })

      it('should split comma-separated string for "nin" operator', () => {
        const entity: ReportEntity = {
          templateId: 'test-template-id',
          organizationId: 'test-org-id',
          filters: {
            fields: [
              {
                database: 'db',
                table: 'tbl',
                field: 'code',
                operator: 'nin',
                values: 'exclude1,exclude2'
              }
            ]
          } as any
        }

        const result = ReporterReportMapper.toCreateDto(entity)

        expect(result.filters.db.tbl.code.nin).toEqual(['exclude1', 'exclude2'])
      })

      it('should split comma-separated string for "between" operator', () => {
        const entity: ReportEntity = {
          templateId: 'test-template-id',
          organizationId: 'test-org-id',
          filters: {
            fields: [
              {
                database: 'db',
                table: 'tbl',
                field: 'amount',
                operator: 'between',
                values: '100,500'
              }
            ]
          } as any
        }

        const result = ReporterReportMapper.toCreateDto(entity)

        expect(result.filters.db.tbl.amount.between).toEqual(['100', '500'])
      })

      it('should trim whitespace from comma-separated values', () => {
        const entity: ReportEntity = {
          templateId: 'test-template-id',
          organizationId: 'test-org-id',
          filters: {
            fields: [
              {
                database: 'db',
                table: 'tbl',
                field: 'status',
                operator: 'in',
                values: '  active  ,  pending  ,  completed  '
              }
            ]
          } as any
        }

        const result = ReporterReportMapper.toCreateDto(entity)

        expect(result.filters.db.tbl.status.in).toEqual([
          'active',
          'pending',
          'completed'
        ])
      })

      it('should filter empty values from comma-separated string', () => {
        const entity: ReportEntity = {
          templateId: 'test-template-id',
          organizationId: 'test-org-id',
          filters: {
            fields: [
              {
                database: 'db',
                table: 'tbl',
                field: 'status',
                operator: 'in',
                values: 'active,,pending,,,completed'
              }
            ]
          } as any
        }

        const result = ReporterReportMapper.toCreateDto(entity)

        expect(result.filters.db.tbl.status.in).toEqual([
          'active',
          'pending',
          'completed'
        ])
      })

      it('should not split string for single-value operators like "eq"', () => {
        const entity: ReportEntity = {
          templateId: 'test-template-id',
          organizationId: 'test-org-id',
          filters: {
            fields: [
              {
                database: 'db',
                table: 'tbl',
                field: 'description',
                operator: 'eq',
                values: 'value with, comma inside'
              }
            ]
          } as any
        }

        const result = ReporterReportMapper.toCreateDto(entity)

        expect(result.filters.db.tbl.description.eq).toEqual([
          'value with, comma inside'
        ])
      })
    })
  })
})
