import { DataSource } from '../entities/data-source-entity'

export abstract class DataSourceRepository {
  abstract fetchAll(): Promise<DataSource[]>
  abstract fetchById(id: string): Promise<DataSource>
}
