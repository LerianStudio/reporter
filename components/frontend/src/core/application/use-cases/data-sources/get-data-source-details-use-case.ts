import { inject, injectable } from 'inversify'
import { LogOperation } from '@/core/infrastructure/logger/decorators/log-operation'
import { DataSourceDetailsDto } from '../../dto/data-source-dto'
import { SmartTemplatesHttpService } from '@/core/infrastructure/smart-templates/services/smart-templates-http-service'

export type GetDataSourceDetailsParams = {
  dataSourceId: string
}

export type GetDataSourceDetails = {
  execute(params: GetDataSourceDetailsParams): Promise<DataSourceDetailsDto>
}

@injectable()
export class GetDataSourceDetailsUseCase implements GetDataSourceDetails {
  constructor(
    @inject(SmartTemplatesHttpService)
    private readonly smartTemplatesHttpService: SmartTemplatesHttpService
  ) {}

  @LogOperation({ layer: 'application' })
  async execute(params: GetDataSourceDetailsParams): Promise<DataSourceDetailsDto> {
    // Fetch data source details from smart templates API
    const response = await this.smartTemplatesHttpService.get<DataSourceDetailsDto>(
      `/v1/data-sources/${params.dataSourceId}`
    )
    
    return response
  }
} 