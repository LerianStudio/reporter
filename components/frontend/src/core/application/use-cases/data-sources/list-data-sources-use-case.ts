import { inject, injectable } from 'inversify'
import { LogOperation } from '@/core/infrastructure/logger/decorators/log-operation'
import { DataSourceInformationDto } from '../../dto/data-source-dto'
import { SmartTemplatesHttpService } from '@/core/infrastructure/smart-templates/services/smart-templates-http-service'

export type ListDataSources = {
  execute(): Promise<DataSourceInformationDto[]>
}

@injectable()
export class ListDataSourcesUseCase implements ListDataSources {
  constructor(
    @inject(SmartTemplatesHttpService)
    private readonly smartTemplatesHttpService: SmartTemplatesHttpService
  ) {}

  @LogOperation({ layer: 'application' })
  async execute(): Promise<DataSourceInformationDto[]> {
    // Fetch all data sources from smart templates API
    const response = await this.smartTemplatesHttpService.get<DataSourceInformationDto[]>('/v1/data-sources')
    
    return response
  }
} 