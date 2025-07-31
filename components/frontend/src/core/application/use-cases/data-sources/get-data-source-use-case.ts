import { inject, injectable } from 'inversify'
import { LogOperation } from '@/core/infrastructure/logger/decorators/log-operation'
import { DataSourceDto } from '../../dto/data-source-dto'
import { DataSourceRepository } from '@/core/domain/repositories/data-source-repository'

export type GetDataSource = {
  execute(id: string): Promise<DataSourceDto>
}

@injectable()
export class GetDataSourceUseCase implements GetDataSource {
  constructor(
    @inject(DataSourceRepository)
    private readonly dataSourceRepository: DataSourceRepository
  ) {}

  @LogOperation({ layer: 'application' })
  async execute(id: string): Promise<DataSourceDto> {
    return this.dataSourceRepository.fetchById(id)
  }
}
