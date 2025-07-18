import { inject, injectable } from 'inversify'
import { ReportRepository } from '@/core/domain/repositories/report-repository'
import { ReportDto } from '../../dto/report-dto'
import { ReportMapper } from '../../mappers/report-mapper'
import { LogOperation } from '@/core/infrastructure/logger/decorators/log-operation'

export type GetReportStatusParams = {
  id: string
  organizationId: string
}

export type GetReportStatus = {
  execute(params: GetReportStatusParams): Promise<ReportDto>
}

@injectable()
export class GetReportStatusUseCase implements GetReportStatus {
  constructor(
    @inject(ReportRepository)
    private readonly reportRepository: ReportRepository
  ) {}

  @LogOperation({ layer: 'application' })
  async execute(params: GetReportStatusParams): Promise<ReportDto> {
    // Fetch report by ID with organization validation
    const report = await this.reportRepository.fetchById(
      params.id,
      params.organizationId
    )

    // Return mapped DTO with current status
    return ReportMapper.toResponseDto(report)
  }
}
