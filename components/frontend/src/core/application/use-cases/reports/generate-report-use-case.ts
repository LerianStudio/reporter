import { inject, injectable } from 'inversify'
import { ReportRepository } from '@/core/domain/repositories/report-repository'
import { type CreateReportDto, type ReportDto } from '../../dto/report-dto'
import { ReportMapper } from '../../mappers/report-mapper'
import { LogOperation } from '@/core/infrastructure/logger/decorators/log-operation'

export type GenerateReport = {
  execute(
    reportData: CreateReportDto & { organizationId: string }
  ): Promise<ReportDto>
}

@injectable()
export class GenerateReportUseCase implements GenerateReport {
  constructor(
    @inject(ReportRepository)
    private readonly reportRepository: ReportRepository
  ) {}

  @LogOperation({ layer: 'application' })
  async execute(
    reportData: CreateReportDto & { organizationId: string }
  ): Promise<ReportDto> {
    const { organizationId, ...dto } = reportData
    const reportEntity = ReportMapper.toEntity(dto, organizationId)

    const createdReport = await this.reportRepository.create(reportEntity)

    return ReportMapper.toResponseDto(createdReport)
  }
}
