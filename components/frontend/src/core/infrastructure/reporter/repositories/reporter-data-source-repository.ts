import { inject, injectable } from 'inversify'
import { DataSourceRepository } from '@/core/domain/repositories/data-source-repository'
import { DataSource } from '@/core/domain/entities/data-source-entity'
import { ReporterHttpService } from '../services/reporter-http-service'
import { ReporterDataSourceMapper } from '../mappers/reporter-data-source-mapper'
import { ReporterDataSourceDto } from '../dto/reporter-data-source-dto'

@injectable()
export class ReporterDataSourceRepository implements DataSourceRepository {
  constructor(
    @inject(ReporterHttpService)
    private readonly reporterHttpService: ReporterHttpService
  ) {}

  async fetchAll(organizationId: string): Promise<DataSource[]> {
    const response = await this.reporterHttpService.get<
      ReporterDataSourceDto[]
    >('/v1/data-sources', {
      headers: {
        'X-Organization-Id': organizationId
      }
    })

    return ReporterDataSourceMapper.toListEntity(response)
  }

  async fetchById(organizationId: string, id: string): Promise<DataSource> {
    const response = await this.reporterHttpService.get<ReporterDataSourceDto>(
      `/v1/data-sources/${id}`,
      {
        headers: {
          'X-Organization-Id': organizationId
        }
      }
    )

    return ReporterDataSourceMapper.toEntity(response)
  }
}
