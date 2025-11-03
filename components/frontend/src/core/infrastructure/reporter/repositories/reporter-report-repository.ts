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
import { TemplateRepository } from '@/core/domain/repositories/template-repository'
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
    private readonly httpService: ReporterHttpService,
    @inject(TemplateRepository)
    private readonly templateRepository: TemplateRepository
  ) {}

  private baseUrl: string = '/v1/reports'

  async create(report: ReportEntity): Promise<ReportEntity> {
    const createDto = ReporterReportMapper.toCreateDto(report)

    const response = await this.httpService.post<ReporterReportDto>(
      this.baseUrl,
      {
        body: JSON.stringify(createDto),
        headers: {
          'X-Organization-Id': report.organizationId,
          'Content-Type': 'application/json'
        }
      }
    )

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

    const template = await this.templateRepository.fetchById(
      report.templateId,
      organizationId
    )

    const outputFormat = template.outputFormat || 'txt'

    // Determine if the format is binary or text
    // Only PDF is binary among the supported formats
    const isBinary = outputFormat.toLowerCase() === 'pdf'

    let content: string | ArrayBuffer

    if (isBinary) {
      // Use getBinary for PDF to preserve binary data integrity
      content = await this.httpService.getBinary(
        `${this.baseUrl}/${id}/download`,
        {
          headers: {
            'X-Organization-Id': organizationId
          }
        }
      )
    } else {
      // Use getText for text-based formats (HTML, XML, CSV, TXT)
      content = await this.httpService.getText(
        `${this.baseUrl}/${id}/download`,
        {
          headers: {
            'X-Organization-Id': organizationId
          }
        }
      )
    }

    const contentTypeMap: Record<string, string> = {
      pdf: 'application/pdf',
      html: 'text/html',
      xml: 'text/xml',
      csv: 'text/csv',
      txt: 'text/plain'
    }

    const contentType =
      contentTypeMap[outputFormat.toLowerCase()] || 'text/plain'

    const sanitizedName = template.name.toLowerCase().replace(/\s+/g, '_')
    const fileName = `${sanitizedName}_${new Date().toISOString().split('T')[0]}.${outputFormat}`

    return {
      content,
      fileName,
      contentType
    }
  }
}
