import { injectable, inject } from 'inversify'
import {
  ReportEntity,
  ReportSearchEntity
} from '@/core/domain/entities/report-entity'
import { PaginationEntity } from '@/core/domain/entities/pagination-entity'
import {
  ReportRepository,
  DownloadFileResponse
} from '@/core/domain/repositories/report-repository'
import { ReporterHttpService } from '../services/reporter-http-service'
import { ReporterReportMapper } from '../mappers/reporter-report-mapper'
import { ReporterReportDto } from '../dto/reporter-report-dto'
import { ReporterPaginationDto } from '../dto/reporter-pagination-dto'
import { createQueryString } from '@/lib/search'
import { validateAndFormatDateForQuery } from '@/lib/date-validation'

@injectable()
export class ReporterReportRepository implements ReportRepository {
  constructor(
    @inject(ReporterHttpService)
    private readonly httpService: ReporterHttpService
  ) {}

  private baseUrl: string = '/v1/reports'

  async create(report: ReportEntity): Promise<ReportEntity> {
    const createDto = ReporterReportMapper.toCreateDto(report)

    const response = await this.httpService.post<ReporterReportDto>(this.baseUrl, {
      body: JSON.stringify(createDto),
      headers: {
        'X-Organization-Id': report.organizationId,
        'Content-Type': 'application/json'
      }
    })

    return {
      ...ReporterReportMapper.toEntity(response),
      organizationId: report.organizationId
    }
  }

  async fetchAll(
    organizationId: string,
    query: ReportSearchEntity
  ): Promise<PaginationEntity<ReportEntity>> {
    const queryParams: Record<string, any> = {
      limit: query.limit,
      page: query.page
    }

    if (query.status) {
      queryParams.status = query.status
    }
    if (query.search) {
      queryParams.search = query.search
    }
    if (query.templateId) {
      queryParams.templateId = query.templateId
    }
    if (query.createdAt) {
      queryParams.createdAt = validateAndFormatDateForQuery(query.createdAt)
    }

    const response = await this.httpService.get<
      ReporterPaginationDto<ReporterReportDto>
    >(`${this.baseUrl}${createQueryString(queryParams)}`, {
      headers: {
        'X-Organization-Id': organizationId
      }
    })

    const paginationResult = ReporterReportMapper.toPaginationEntity(response)

    return {
      ...paginationResult,
      items: paginationResult.items.map((report) => ({
        ...report,
        organizationId
      }))
    }
  }

  async fetchById(id: string, organizationId: string): Promise<ReportEntity> {
    const response = await this.httpService.get<ReporterReportDto>(
      `${this.baseUrl}/${id}`,
      {
        headers: {
          'X-Organization-Id': organizationId
        }
      }
    )

    return {
      ...ReporterReportMapper.toEntity(response),
      organizationId
    }
  }

  async downloadFile(
    id: string,
    organizationId: string
  ): Promise<DownloadFileResponse> {
    const report = await this.fetchById(id, organizationId)

    if (report.status !== 'Finished') {
      throw new Error(
        `Report is not ready for download. Current status: ${report.status}`
      )
    }

    const content = await this.httpService.getText(
      `${this.baseUrl}/${id}/download`,
      {
        headers: {
          'X-Organization-Id': organizationId
        }
      }
    )

    const fileName = `report_${id}_${new Date().toISOString().split('T')[0]}.txt`

    const contentType = 'text/plain'

    return {
      content,
      fileName,
      contentType
    }
  }
}
