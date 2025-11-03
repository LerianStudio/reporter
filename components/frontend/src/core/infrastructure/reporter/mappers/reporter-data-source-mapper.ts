import { ReporterDataSourceDto } from '../dto/reporter-data-source-dto'
import { DataSource } from '@/core/domain/entities/data-source-entity'

export class ReporterDataSourceMapper {
  static toEntity(dto: ReporterDataSourceDto): DataSource {
    return {
      id: dto.id,
      externalName: dto.externalName,
      type: dto.type,
      tables: dto.tables?.map((table) => ({
        name: table.name,
        fields: table.fields.map((field) =>
          typeof field === 'string' ? { name: field, type: 'string' } : field
        )
      }))
    }
  }

  static toListEntity(dto: ReporterDataSourceDto[]): DataSource[] {
    return (
      dto
        .filter((item) => item.id !== '')
        .map(ReporterDataSourceMapper.toEntity) ?? []
    )
  }
}
