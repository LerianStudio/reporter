import { NextResponse } from 'next/server'
import { z } from 'zod'
import { GenerateReportUseCase } from '../use-cases/reports/generate-report-use-case'
import { ListReportsUseCase } from '../use-cases/reports/list-reports-use-case'
import { GetReportStatusUseCase } from '../use-cases/reports/get-report-status-use-case'
import { DownloadReportUseCase } from '../use-cases/reports/download-report-use-case'
import type { ReportSearchParamDto } from '../dto/report-dto'
import {
  Body,
  Controller,
  Get,
  Inject,
  Param,
  Post,
  Query
} from '@lerianstudio/sindarian-server'

const FilterFieldSchema = z.object({
  database: z.string(),
  table: z.string(),
  field: z.string(),
  operator: z.enum([
    'eq',
    'gt',
    'gte',
    'lt',
    'lte',
    'between',
    'in',
    'nin'
  ] as const),
  values: z.union([z.string(), z.array(z.string())])
})

const CreateReportSchema = z.object({
  templateId: z.string().uuid('Template ID must be a valid UUID'),
  fields: z.array(FilterFieldSchema).default([])
})

type CreateReportData = z.infer<typeof CreateReportSchema>

@Controller('/organizations/:id/reports')
export class ReportController {
  constructor(
    @Inject(GenerateReportUseCase)
    private readonly generateReportUseCase: GenerateReportUseCase,
    @Inject(ListReportsUseCase)
    private readonly listReportsUseCase: ListReportsUseCase,
    @Inject(GetReportStatusUseCase)
    private readonly getReportStatusUseCase: GetReportStatusUseCase,
    @Inject(DownloadReportUseCase)
    private readonly downloadReportUseCase: DownloadReportUseCase
  ) {}

  @Get('/:reportId')
  async fetchById(
    @Param('id') organizationId: string,
    @Param('reportId') reportId: string
  ) {
    return await this.getReportStatusUseCase.execute({
      id: reportId!,
      organizationId
    })
  }

  @Get('/')
  async fetchAll(
    @Param('id') organizationId: string,
    @Query() query: ReportSearchParamDto
  ) {
    return await this.listReportsUseCase.execute(organizationId, query)
  }

  @Post('/')
  async create(
    @Param('id') organizationId: string,
    @Body() body: CreateReportData
  ) {
    const report = await this.generateReportUseCase.execute({
      templateId: body.templateId,
      organizationId,
      fields: body.fields
    })

    return NextResponse.json(report, { status: 201 })
  }

  @Get('/:reportId/download')
  async download(
    @Param('id') organizationId: string,
    @Param('reportId') reportId: string
  ) {
    try {
      if (!reportId) {
        return NextResponse.json(
          { error: 'Report ID is required' },
          { status: 400 }
        )
      }

      const downloadInfo = await this.downloadReportUseCase.execute({
        id: reportId,
        organizationId
      })

      return new NextResponse(downloadInfo.content, {
        status: 200,
        headers: {
          'Content-Type': downloadInfo.contentType,
          'Content-Disposition': `attachment; filename="${downloadInfo.fileName}"`
        }
      })
    } catch (error) {
      console.error('Report download error:', error)

      const errorMessage =
        error instanceof Error ? error.message : 'Failed to download report'

      if (
        errorMessage.includes('not ready for download') ||
        errorMessage.includes('not available')
      ) {
        return NextResponse.json({ error: errorMessage }, { status: 409 })
      }

      return NextResponse.json(
        { error: 'Failed to download report' },
        { status: 500 }
      )
    }
  }

  @Get('/:reportId/download-info')
  async getDownloadInfo(
    @Param('id') organizationId: string,
    @Param('reportId') reportId: string
  ) {
    try {
      if (!reportId) {
        return NextResponse.json(
          { error: 'Report ID is required' },
          { status: 400 }
        )
      }

      const downloadInfo = await this.downloadReportUseCase.execute({
        id: reportId,
        organizationId
      })

      return NextResponse.json({
        fileName: downloadInfo.fileName,
        contentType: downloadInfo.contentType,
        isReady: true
      })
    } catch (error) {
      console.error('Report download info error:', error)

      const errorMessage =
        error instanceof Error ? error.message : 'Failed to get download info'

      if (
        errorMessage.includes('not ready for download') ||
        errorMessage.includes('not available')
      ) {
        return NextResponse.json(
          {
            error: errorMessage,
            isReady: false
          },
          { status: 409 }
        )
      }

      return NextResponse.json(
        { error: 'Failed to get download info' },
        { status: 500 }
      )
    }
  }
}
