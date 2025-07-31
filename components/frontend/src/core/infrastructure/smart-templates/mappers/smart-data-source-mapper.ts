import { SmartDataSourceDto } from '../dto/smart-data-source-dto'
import { DataSource } from '@/core/domain/entities/data-source-entity'

export class SmartDataSourceMapper {
  static toEntity(dto: SmartDataSourceDto): DataSource {
    return {
      id: dto.id,
      externalName: dto.externalName,
      type: dto.type,
      tables: dto.tables?.map((table) => ({
        name: table.name,
        fields: table.fields
      }))
    }
  }
}
