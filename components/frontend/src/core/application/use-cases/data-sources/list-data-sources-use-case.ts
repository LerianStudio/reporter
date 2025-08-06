import { inject, injectable } from 'inversify'
import { LogOperation } from '@/core/infrastructure/logger/decorators/log-operation'
import { DataSourceDto } from '../../dto/data-source-dto'
import { DataSourceRepository } from '@/core/domain/repositories/data-source-repository'

export type ListDataSources = {
  execute(organizationId: string): Promise<DataSourceDto[]>
}

@injectable()
export class ListDataSourcesUseCase implements ListDataSources {
  constructor(
    @inject(DataSourceRepository)
    private readonly dataSourceRepository: DataSourceRepository
  ) {}

  @LogOperation({ layer: 'application' })
  async execute(organizationId: string): Promise<DataSourceDto[]> {
    return this.dataSourceRepository.fetchAll(organizationId)
  }
}
