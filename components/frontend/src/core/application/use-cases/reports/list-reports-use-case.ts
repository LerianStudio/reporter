import { inject, injectable } from 'inversify'
import { ReportRepository } from '@/core/domain/repositories/report-repository'
import { TemplateRepository } from '@/core/domain/repositories/template-repository'
import type { ReportSearchEntity } from '@/core/domain/entities/report-entity'
import type { ReportDto, ReportSearchParamDto } from '../../dto/report-dto'
import { ReportMapper } from '../../mappers/report-mapper'
import { LogOperation } from '@/core/infrastructure/logger/decorators/log-operation'
import { PaginationDto } from '../../dto/pagination-dto'

export type ListReports = {
  execute(
    organizationId: string,
    query: ReportSearchParamDto
  ): Promise<PaginationDto<ReportDto>>
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
  async execute(
    organizationId: string,
    query: ReportSearchParamDto
  ): Promise<PaginationDto<ReportDto>> {
    let createdAtDate: Date | undefined = undefined
    if (query.createdAt) {
      const date = new Date(query.createdAt)
      createdAtDate = isNaN(date.getTime()) ? undefined : date
    }

    const searchEntity: ReportSearchEntity = {
      ...query,
      createdAt: createdAtDate
    }

    const reports = await this.reportRepository.fetchAll(
      organizationId,
      searchEntity
    )

    const reportsWithTemplates = await Promise.allSettled(
      reports.items.map(async (report) => {
        try {
          const template = await this.templateRepository.fetchById(
            report.templateId,
            organizationId
          )
          return {
            ...report,
            template
          }
        } catch (error) {
          console.warn(
            `Failed to fetch template ${report.templateId} for report ${report.id}:`,
            error
          )
          return report
        }
      })
    )

    const enrichedReports = reportsWithTemplates.map((result) => {
      if (result.status === 'fulfilled') {
        return result.value
      } else {
        console.warn('Unexpected template fetch failure:', result.reason)
        return result.reason
      }
    })

    const enrichedPagination = {
      ...reports,
      items: enrichedReports
    }

    return ReportMapper.toPaginationResponseDto(enrichedPagination)
  }
}
