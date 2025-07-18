import { injectable, inject } from 'inversify'
import {
  ReportEntity,
  ReportStatus
} from '@/core/domain/entities/report-entity'
import { PaginationEntity } from '@/core/domain/entities/pagination-entity'
import {
  ReportRepository,
  ReportQueryFilters,
  FetchReportsParams,
  DownloadFileResponse
} from '@/core/domain/repositories/report-repository'
import { SmartTemplatesHttpService } from '../services/smart-templates-http-service'
import { SmartReportMapper } from '../mappers/smart-report-mapper'
import { SmartReportDto } from '../dto/smart-report-dto'
import { SmartPaginationDto } from '../dto/smart-pagination-dto'
import { createQueryString } from '@/lib/search'

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
    params: FetchReportsParams
  ): Promise<PaginationEntity<ReportEntity>> {
    const queryParams: Record<string, any> = {
      limit: params.limit,
      page: params.page
    }

    // Add filters to query params
    if (params.filters) {
      if (params.filters.status) {
        queryParams.status = params.filters.status
      }
      if (params.filters.templateId) {
        queryParams.templateId = params.filters.templateId
      }
      if (params.filters.search) {
        queryParams.search = params.filters.search
      }
    }

    const response = await this.httpService.get<
      SmartPaginationDto<SmartReportDto>
    >(`${this.baseUrl}${createQueryString(queryParams)}`, {
      headers: {
        'X-Organization-Id': params.organizationId
      }
    })

    const paginationResult = SmartReportMapper.toPaginationEntity(response)

    // Set organizationId for all entities
    return {
      ...paginationResult,
      items: paginationResult.items.map((report) => ({
        ...report,
        organizationId: params.organizationId
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

  async fetchByStatus(
    organizationId: string,
    status: ReportStatus,
    limit: number,
    page: number
  ): Promise<PaginationEntity<ReportEntity>> {
    return this.fetchAll({
      organizationId,
      limit,
      page,
      filters: { status }
    })
  }

  async fetchByTemplate(
    organizationId: string,
    templateId: string,
    limit: number,
    page: number
  ): Promise<PaginationEntity<ReportEntity>> {
    return this.fetchAll({
      organizationId,
      limit,
      page,
      filters: { templateId }
    })
  }

  async getDownloadUrl(id: string, organizationId: string): Promise<string> {
    const response = await this.httpService.get<{ downloadUrl: string }>(
      `${this.baseUrl}/${id}/download-url`,
      {
        headers: {
          'X-Organization-Id': organizationId
        }
      }
    )

    return response.downloadUrl
  }

  async downloadFile(
    id: string,
    organizationId: string
  ): Promise<DownloadFileResponse> {
    // First, get the report to ensure it's downloadable and get metadata
    const report = await this.fetchById(id, organizationId)

    // Verify report is in downloadable state
    if (report.status !== 'Finished') {
      throw new Error(
        `Report is not ready for download. Current status: ${report.status}`
      )
    }

    // Call the download endpoint that returns file content directly as text
    const content = await this.httpService.getText(
      `${this.baseUrl}/${id}/download`,
      {
        headers: {
          'X-Organization-Id': organizationId
        }
      }
    )

    // Generate filename based on report info - default to .txt since we don't have URL info
    const fileName = `report_${id}_${new Date().toISOString().split('T')[0]}.txt`

    // Default content type to text/plain since we no longer have URL to infer from
    const contentType = 'text/plain'

    return {
      content,
      fileName,
      contentType
    }
  }

  async countByOrganization(
    organizationId: string,
    filters?: ReportQueryFilters
  ): Promise<number> {
    const queryParams: Record<string, any> = {}

    // Add filters to query params
    if (filters) {
      if (filters.status) {
        queryParams.status = filters.status
      }
      if (filters.templateId) {
        queryParams.templateId = filters.templateId
      }
      // Add other filters as needed
    }

    const response = await this.httpService.get<{ count: number }>(
      `${this.baseUrl}/count${createQueryString(queryParams)}`,
      {
        headers: {
          'X-Organization-Id': organizationId
        }
      }
    )

    return response.count
  }

  async countByStatus(
    organizationId: string,
    status: ReportStatus
  ): Promise<number> {
    return this.countByOrganization(organizationId, { status })
  }

  async search(
    organizationId: string,
    searchText: string,
    limit: number,
    page: number
  ): Promise<PaginationEntity<ReportEntity>> {
    return this.fetchAll({
      organizationId,
      limit,
      page,
      filters: { search: searchText }
    })
  }

  async fetchProcessingReports(
    organizationId: string,
    olderThan?: Date
  ): Promise<ReportEntity[]> {
    const queryParams: Record<string, any> = {
      status: 'Processing'
    }

    if (olderThan) {
      queryParams.createdBefore = olderThan.toISOString()
    }

    const response = await this.httpService.get<SmartReportDto[]>(
      `${this.baseUrl}/processing${createQueryString(queryParams)}`,
      {
        headers: {
          'X-Organization-Id': organizationId
        }
      }
    )

    return response.map((dto) => ({
      ...SmartReportMapper.toEntity(dto),
      organizationId
    }))
  }
}
