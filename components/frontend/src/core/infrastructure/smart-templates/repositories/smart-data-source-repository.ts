import { inject, injectable } from 'inversify'
import { DataSourceRepository } from '@/core/domain/repositories/data-source-repository'
import { DataSource } from '@/core/domain/entities/data-source-entity'
import { SmartTemplatesHttpService } from '../services/smart-templates-http-service'
import { SmartDataSourceMapper } from '../mappers/smart-data-source-mapper'
import { SmartDataSourceDto } from '../dto/smart-data-source-dto'

@injectable()
export class SmartDataSourceRepository implements DataSourceRepository {
  constructor(
    @inject(SmartTemplatesHttpService)
    private readonly smartTemplatesHttpService: SmartTemplatesHttpService
  ) {}

  async fetchAll(): Promise<DataSource[]> {
    // Fetch all data sources from smart templates API
    const response =
      await this.smartTemplatesHttpService.get<SmartDataSourceDto[]>(
        '/v1/data-sources'
      )

    return response?.map(SmartDataSourceMapper.toEntity) || []
  }

  async fetchById(id: string): Promise<DataSource> {
    // Fetch data source by ID from smart templates API
    const response =
      await this.smartTemplatesHttpService.get<SmartDataSourceDto>(
        `/v1/data-sources/${id}`
      )

    return SmartDataSourceMapper.toEntity(response)
  }
}
