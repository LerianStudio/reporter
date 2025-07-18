import { inject, injectable } from 'inversify'
import { type ReportEntity } from '@/core/domain/entities/report-entity'
import { ReportRepository } from '@/core/domain/repositories/report-repository'
import { type CreateReportDto, type ReportDto } from '../../dto/report-dto'
import { ReportMapper } from '../../mappers/report-mapper'
import { LogOperation } from '@/core/infrastructure/logger/decorators/log-operation'

export type GenerateReport = {
  execute(reportData: CreateReportDto): Promise<ReportDto>
}

@injectable()
export class GenerateReportUseCase implements GenerateReport {
  constructor(
    @inject(ReportRepository)
    private readonly reportRepository: ReportRepository
  ) {}

  @LogOperation({ layer: 'application' })
  async execute(reportData: CreateReportDto): Promise<ReportDto> {
    // Create report entity for async processing
    const reportEntity: ReportEntity = {
      templateId: reportData.templateId,
      organizationId: reportData.organizationId,
      filters: reportData.filters
    }

    // Create report request (will be processed asynchronously by backend)
    const createdReport = await this.reportRepository.create(reportEntity)

    // Return mapped DTO
    return ReportMapper.toResponseDto(createdReport)
  }
}
