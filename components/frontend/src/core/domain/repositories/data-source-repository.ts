import { DataSource } from '../entities/data-source-entity'

export abstract class DataSourceRepository {
  abstract fetchAll(organizationId: string): Promise<DataSource[]>
  abstract fetchById(organizationId: string, id: string): Promise<DataSource>
}
