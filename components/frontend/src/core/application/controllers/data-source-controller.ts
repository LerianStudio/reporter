import { ListDataSourcesUseCase } from '../use-cases/data-sources/list-data-sources-use-case'
import { GetDataSourceUseCase } from '../use-cases/data-sources/get-data-source-use-case'
import { Controller, Get, Inject, Param } from '@lerianstudio/sindarian-server'

/**
 * Data Source Controller
 *
 * Next.js API route controller for handling data source-related HTTP requests.
 * Provides RESTful endpoints for retrieving data source information and details.
 * Follows console patterns with proper request/response handling.
 */
@Controller('/organizations/:id/data-sources')
export class DataSourceController {
  constructor(
    @Inject(ListDataSourcesUseCase)
    private readonly listDataSourcesUseCase: ListDataSourcesUseCase,
    @Inject(GetDataSourceUseCase)
    private readonly getDataSourceDetailsUseCase: GetDataSourceUseCase
  ) {}

  /**
   * GET /api/data-sources
   * Fetches all available data sources
   */
  @Get('/')
  async fetchAll(@Param('id') organizationId: string) {
    return await this.listDataSourcesUseCase.execute(organizationId)
  }

  /**
   * GET /api/data-sources/[dataSourceId]
   * Fetches detailed information for a specific data source
   */
  @Get('/:dataSourceId')
  async fetchById(
    @Param('id') organizationId: string,
    @Param('dataSourceId') dataSourceId: string
  ) {
    return await this.getDataSourceDetailsUseCase.execute(
      organizationId,
      dataSourceId!
    )
  }
}
