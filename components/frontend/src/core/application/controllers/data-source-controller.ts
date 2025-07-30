import { inject, injectable } from 'inversify'
import { NextResponse } from 'next/server'
import { Controller } from '@/lib/http/server/decorators/controller-decorator'
import { LoggerInterceptor } from '@/core/infrastructure/logger/decorators'
import { ListDataSourcesUseCase } from '../use-cases/data-sources/list-data-sources-use-case'
import { GetDataSourceUseCase } from '../use-cases/data-sources/get-data-source-use-case'

type DataSourceParams = {
  dataSourceId?: string
}

/**
 * Data Source Controller
 *
 * Next.js API route controller for handling data source-related HTTP requests.
 * Provides RESTful endpoints for retrieving data source information and details.
 * Follows console patterns with proper request/response handling.
 */
@injectable()
@LoggerInterceptor()
@Controller()
export class DataSourceController {
  constructor(
    @inject(ListDataSourcesUseCase)
    private readonly listDataSourcesUseCase: ListDataSourcesUseCase,
    @inject(GetDataSourceUseCase)
    private readonly getDataSourceDetailsUseCase: GetDataSourceUseCase
  ) {}

  /**
   * GET /api/data-sources
   * Fetches all available data sources
   */
  async fetchAll(): Promise<NextResponse> {
    const dataSources = await this.listDataSourcesUseCase.execute()

    return NextResponse.json(dataSources, { status: 200 })
  }

  /**
   * GET /api/data-sources/[dataSourceId]
   * Fetches detailed information for a specific data source
   */
  async fetchById(
    request: Request,
    { params }: { params: DataSourceParams }
  ): Promise<NextResponse> {
    const { dataSourceId } = await params

    const dataSourceDetails = await this.getDataSourceDetailsUseCase.execute(
      dataSourceId!
    )

    return NextResponse.json(dataSourceDetails, { status: 200 })
  }
}
