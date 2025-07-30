import { inject, injectable } from 'inversify'
import { NextResponse } from 'next/server'
import { z } from 'zod'
import { Controller } from '@/lib/http/server/decorators/controller-decorator'
import { LoggerInterceptor } from '@/core/infrastructure/logger/decorators'
import { ListDataSourcesUseCase } from '../use-cases/data-sources/list-data-sources-use-case'
import { GetDataSourceDetailsUseCase } from '../use-cases/data-sources/get-data-source-details-use-case'

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
    @inject(GetDataSourceDetailsUseCase)
    private readonly getDataSourceDetailsUseCase: GetDataSourceDetailsUseCase
  ) {}

  /**
   * GET /api/data-sources
   * Fetches all available data sources
   */
  async fetchAll(request: Request): Promise<NextResponse> {
    try {
      const dataSources = await this.listDataSourcesUseCase.execute()

      return NextResponse.json(dataSources, { status: 200 })
    } catch (error) {
      console.error('[ERROR] - DataSourceController.fetchAll', error)
      return NextResponse.json(
        { error: 'Failed to fetch data sources' },
        { status: 500 }
      )
    }
  }

  /**
   * GET /api/data-sources/[dataSourceId]
   * Fetches detailed information for a specific data source
   */
  async fetchById(request: Request, params: { params: DataSourceParams }): Promise<NextResponse> {
    try {
      const { dataSourceId } = params.params

      if (!dataSourceId) {
        return NextResponse.json(
          { error: 'Data source ID is required' },
          { status: 400 }
        )
      }

      const dataSourceDetails = await this.getDataSourceDetailsUseCase.execute({
        dataSourceId
      })

      return NextResponse.json(dataSourceDetails, { status: 200 })
    } catch (error) {
      console.error('[ERROR] - DataSourceController.fetchById', error)
      return NextResponse.json(
        { error: 'Failed to fetch data source details' },
        { status: 500 }
      )
    }
  }
} 