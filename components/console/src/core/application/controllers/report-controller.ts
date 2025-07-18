import { inject, injectable } from 'inversify'
import { NextResponse } from 'next/server'
import { z } from 'zod'
import { Controller } from '@/lib/http/server/decorators/controller-decorator'
import { LoggerInterceptor } from '@/core/infrastructure/logger/decorators'
import { GenerateReportUseCase } from '../use-cases/reports/generate-report-use-case'
import { ListReportsUseCase } from '../use-cases/reports/list-reports-use-case'
import { GetReportStatusUseCase } from '../use-cases/reports/get-report-status-use-case'
import { DownloadReportUseCase } from '../use-cases/reports/download-report-use-case'
import { ValidateFormData } from '@/lib/zod/decorators/validate-form-data'
import { ReportStatus } from '@/core/domain/entities/report-entity'

type ReportParams = {
  id: string // organization ID
  reportId?: string
}

const CreateReportSchema = z.object({
  templateId: z.string().uuid('Template ID must be a valid UUID'),
  organizationId: z.string().min(1, 'Organization ID is required'),
  filters: z
    .object({
      ledger_ids: z.array(z.string().uuid()).optional(),
      date_range: z
        .object({
          start: z.string().datetime('Invalid start date format'),
          end: z.string().datetime('Invalid end date format')
        })
        .refine(
          (data) => new Date(data.start) <= new Date(data.end),
          'Start date must be before or equal to end date'
        )
        .optional(),
      account_types: z.array(z.string()).optional(),
      minimum_balance: z.number().min(0).optional(),
      maximum_balance: z.number().min(0).optional(),
      asset_codes: z.array(z.string()).optional(),
      portfolio_ids: z.array(z.string().uuid()).optional(),
      search: z.string().max(255).optional()
    })
    .optional()
})

/**
 * Report Controller
 *
 * Next.js API route controller for handling report-related HTTP requests.
 * Provides RESTful endpoints for report generation, status tracking, and file downloads.
 * Supports async processing, status polling, and streaming file downloads.
 * Follows console patterns with proper request/response handling.
 */
@injectable()
@LoggerInterceptor()
@Controller()
export class ReportController {
  constructor(
    @inject(GenerateReportUseCase)
    private readonly generateReportUseCase: GenerateReportUseCase,
    @inject(ListReportsUseCase)
    private readonly listReportsUseCase: ListReportsUseCase,
    @inject(GetReportStatusUseCase)
    private readonly getReportStatusUseCase: GetReportStatusUseCase,
    @inject(DownloadReportUseCase)
    private readonly downloadReportUseCase: DownloadReportUseCase
  ) {}

  /**
   * Get a specific report status by ID
   * GET /api/organizations/{id}/reports/{reportId}
   */
  async fetchById(request: Request, { params }: { params: ReportParams }) {
    const { id: organizationId, reportId } = await params

    const report = await this.getReportStatusUseCase.execute({
      id: reportId!,
      organizationId
    })

    return NextResponse.json(report)
  }

  /**
   * List reports with pagination and filtering
   * GET /api/organizations/{id}/reports
   */
  async fetchAll(request: Request, { params }: { params: ReportParams }) {
    const { searchParams } = new URL(request.url)
    const limit = Number(searchParams.get('limit')) || 10
    const page = Number(searchParams.get('page')) || 1
    const status = searchParams.get('status') as ReportStatus | undefined
    const search = searchParams.get('search') || undefined

    const { id: organizationId } = await params

    const reports = await this.listReportsUseCase.execute({
      organizationId,
      limit,
      page,
      status,
      search
    })

    return NextResponse.json(reports)
  }

  /**
   * Generate a new report (async processing)
   * POST /api/organizations/{id}/reports
   */
  async create(request: Request) {
    const requestData = await request.json()

    const report = await this.generateReportUseCase.execute({
      templateId: requestData.templateId,
      organizationId: requestData.organizationId,
      filters: requestData.filters
    })

    return NextResponse.json(report, { status: 201 })
  }

  /**
   * Download completed report file (streaming)
   * GET /api/organizations/{id}/reports/{reportId}/download
   */
  async download(request: Request, { params }: { params: ReportParams }) {
    try {
      const organizationId = (await params).id
      const reportId = (await params).reportId

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
  async getDownloadInfo(
    request: Request,
    { params }: { params: ReportParams }
  ) {
    try {
      const organizationId = (await params).id
      const reportId = (await params).reportId

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
