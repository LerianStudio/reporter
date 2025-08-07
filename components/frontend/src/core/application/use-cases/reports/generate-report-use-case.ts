import { inject, injectable } from 'inversify'
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
    const reportEntity = ReportMapper.toEntity(reportData)

    const createdReport = await this.reportRepository.create(reportEntity)

    return ReportMapper.toResponseDto(createdReport)
  }
}
