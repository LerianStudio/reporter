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

const CreateReportSchema = z.object({
  templateId: z.string().uuid('Template ID must be a valid UUID'),
  organizationId: z.string().min(1, 'Organization ID is required'),
  filters: z
    .object({
      fields: z
        .array(
          z.object({
            database: z.string(),
            table: z.string(),
            field: z.string(),
            values: z.array(z.string())
          })
        )
        .optional()
    })
    .optional()
})

type CreateReportData = z.infer<typeof CreateReportSchema>

/**
 * Report Controller
 *
 * Next.js API route controller for handling report-related HTTP requests.
 * Provides RESTful endpoints for report generation, status tracking, and file downloads.
 * Supports async processing, status polling, and streaming file downloads.
 * Follows console patterns with proper request/response handling.
 */
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

  /**
   * Get a specific report status by ID
   * GET /api/organizations/{id}/reports/{reportId}
   */
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

  /**
   * List reports with pagination and filtering
   * GET /api/organizations/{id}/reports
   */
  @Get('/')
  async fetchAll(
    @Param('id') organizationId: string,
    @Query() query: ReportSearchParamDto
  ) {
    return await this.listReportsUseCase.execute(organizationId, query)
  }

  /**
   * Generate a new report (async processing)
   * POST /api/organizations/{id}/reports
   */
  @Post('/')
  async create(
    @Param('id') organizationId: string,
    @Body() body: CreateReportData
  ) {
    const report = await this.generateReportUseCase.execute({
      templateId: body.templateId,
      organizationId,
      filters: body.filters
    })

    return NextResponse.json(report, { status: 201 })
  }

  /**
   * Download completed report file (streaming)
   * GET /api/organizations/{id}/reports/{reportId}/download
   */
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

      // Return file content with appropriate headers for file download
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

      // Check if it's a validation/business error
      if (
        errorMessage.includes('not ready for download') ||
        errorMessage.includes('not available')
      ) {
        return NextResponse.json(
          { error: errorMessage },
          { status: 409 } // Conflict - report not ready
        )
      }

      return NextResponse.json(
        { error: 'Failed to download report' },
        { status: 500 }
      )
    }
  }

  /**
   * Get download information for a report (without streaming)
   * GET /api/organizations/{id}/reports/{reportId}/download-info
   */
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

      // Return download information without providing the actual content
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
