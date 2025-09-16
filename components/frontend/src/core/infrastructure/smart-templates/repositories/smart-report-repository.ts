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
import { SmartTemplatesHttpService } from '../services/smart-templates-http-service'
import { SmartReportMapper } from '../mappers/smart-report-mapper'
import { SmartReportDto } from '../dto/smart-report-dto'
import { SmartPaginationDto } from '../dto/smart-pagination-dto'
import { createQueryString } from '@/lib/search'
import { validateAndFormatDateForQuery } from '@/lib/date-validation'

@injectable()
export class SmartReportRepository implements ReportRepository {
  constructor(
    @inject(SmartTemplatesHttpService)
    private readonly httpService: SmartTemplatesHttpService
  ) {}

  private baseUrl: string = '/v1/reports'

  async create(report: ReportEntity): Promise<ReportEntity> {
    const createDto = SmartReportMapper.toCreateDto(report)

    const response = await this.httpService.post<SmartReportDto>(this.baseUrl, {
      body: JSON.stringify(createDto),
      headers: {
        'X-Organization-Id': report.organizationId,
        'Content-Type': 'application/json'
      }
    })

    return {
      ...SmartReportMapper.toEntity(response),
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
      SmartPaginationDto<SmartReportDto>
    >(`${this.baseUrl}${createQueryString(queryParams)}`, {
      headers: {
        'X-Organization-Id': organizationId
      }
    })

    const paginationResult = SmartReportMapper.toPaginationEntity(response)

    return {
      ...paginationResult,
      items: paginationResult.items.map((report) => ({
        ...report,
        organizationId
      }))
    }
  }

  async fetchById(id: string, organizationId: string): Promise<ReportEntity> {
    const response = await this.httpService.get<SmartReportDto>(
      `${this.baseUrl}/${id}`,
      {
        headers: {
          'X-Organization-Id': organizationId
        }
      }
    )

    return {
      ...SmartReportMapper.toEntity(response),
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
