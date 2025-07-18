import { inject, injectable } from 'inversify'
import {
  ReportRepository,
  ReportQueryFilters
} from '@/core/domain/repositories/report-repository'
import { TemplateRepository } from '@/core/domain/repositories/template-repository'
import { ReportDto } from '../../dto/report-dto'
import { ReportMapper } from '../../mappers/report-mapper'
import { LogOperation } from '@/core/infrastructure/logger/decorators/log-operation'
import { PaginationDto } from '../../dto/pagination-dto'
import { ReportStatus } from '@/core/domain/entities/report-entity'

export type ListReportsParams = {
  organizationId: string
  limit: number
  page: number
  status?: ReportStatus
  search?: string
}

export type ListReports = {
  execute(params: ListReportsParams): Promise<PaginationDto<ReportDto>>
}

@injectable()
export class ListReportsUseCase implements ListReports {
  constructor(
    @inject(ReportRepository)
    private readonly reportRepository: ReportRepository,
    @inject(TemplateRepository)
    private readonly templateRepository: TemplateRepository
  ) {}

  @LogOperation({ layer: 'application' })
  async execute(params: ListReportsParams): Promise<PaginationDto<ReportDto>> {
    // Build filters from parameters
    const filters: ReportQueryFilters = {
      status: params.status,
      search: params.search
    }

    // Fetch paginated reports
    const reports = await this.reportRepository.fetchAll({
      organizationId: params.organizationId,
      limit: params.limit,
      page: params.page,
      filters
    })

    // Fetch template data for each report
    const reportsWithTemplates = await Promise.allSettled(
      reports.items.map(async (report) => {
        try {
          const template = await this.templateRepository.fetchById(
            report.templateId,
            params.organizationId
          )
          return {
            ...report,
            template
          }
        } catch (error) {
          // If template fetch fails, return report without template data
          // This ensures the list still works even if some templates are missing
          console.warn(
            `Failed to fetch template ${report.templateId} for report ${report.id}:`,
            error
          )
          return report
        }
      })
    )

    // Extract successful results and handle failed template fetches gracefully
    const enrichedReports = reportsWithTemplates.map((result) => {
      if (result.status === 'fulfilled') {
        return result.value
      } else {
        // This shouldn't happen due to try-catch, but adding as safety
        console.warn('Unexpected template fetch failure:', result.reason)
        return result.reason // This would be the original report from the catch block
      }
    })

    // Create new pagination result with enriched reports
    const enrichedPagination = {
      ...reports,
      items: enrichedReports
    }

    // Map to DTO
    return ReportMapper.toPaginationResponseDto(enrichedPagination)
  }
}
