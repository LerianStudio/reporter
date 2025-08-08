import { inject } from 'inversify'
import { NextResponse } from 'next/server'
import { Controller } from '@/lib/http/server/decorators/controller-decorator'
import { LoggerInterceptor } from '@/core/infrastructure/logger/decorators'
import { ListDataSourcesUseCase } from '../use-cases/data-sources/list-data-sources-use-case'
import { GetDataSourceUseCase } from '../use-cases/data-sources/get-data-source-use-case'
import { Get, Param } from '@/lib/http/server'
import { BaseController } from '@/lib/http/server/base-controller'

/**
 * Data Source Controller
 *
 * Next.js API route controller for handling data source-related HTTP requests.
 * Provides RESTful endpoints for retrieving data source information and details.
 * Follows console patterns with proper request/response handling.
 */
@LoggerInterceptor()
@Controller()
export class DataSourceController extends BaseController {
  constructor(
    @inject(ListDataSourcesUseCase)
    private readonly listDataSourcesUseCase: ListDataSourcesUseCase,
    @inject(GetDataSourceUseCase)
    private readonly getDataSourceDetailsUseCase: GetDataSourceUseCase
  ) {
    super()
  }

  /**
   * GET /api/data-sources
   * Fetches all available data sources
   */
  @Get()
  async fetchAll(@Param('id') organizationId: string): Promise<NextResponse> {
    const dataSources =
      await this.listDataSourcesUseCase.execute(organizationId)

    return NextResponse.json(dataSources, { status: 200 })
  }

  /**
   * GET /api/data-sources/[dataSourceId]
   * Fetches detailed information for a specific data source
   */
  @Get()
  async fetchById(
    @Param('id') organizationId: string,
    @Param('dataSourceId') dataSourceId: string
  ): Promise<NextResponse> {
    const dataSourceDetails = await this.getDataSourceDetailsUseCase.execute(
      organizationId,
      dataSourceId!
    )

    return NextResponse.json(dataSourceDetails, { status: 200 })
  }
}
