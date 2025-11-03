import { inject, injectable } from 'inversify'
import { ReportRepository } from '@/core/domain/repositories/report-repository'
import { LogOperation } from '@/core/infrastructure/logger/decorators/log-operation'

export type DownloadReportParams = {
  id: string
  organizationId: string
}

export type DownloadReportResponse = {
  content: string | ArrayBuffer
  fileName: string
  contentType: string
}

export type DownloadReport = {
  execute(params: DownloadReportParams): Promise<DownloadReportResponse>
}

@injectable()
export class DownloadReportUseCase implements DownloadReport {
  constructor(
    @inject(ReportRepository)
    private readonly reportRepository: ReportRepository
  ) {}

  @LogOperation({ layer: 'application' })
  async execute(params: DownloadReportParams): Promise<DownloadReportResponse> {
    const downloadInfo = await this.reportRepository.downloadFile(
      params.id,
      params.organizationId
    )

    return {
      content: downloadInfo.content,
      fileName: downloadInfo.fileName,
      contentType: downloadInfo.contentType
    }
  }
}
