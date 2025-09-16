export type DataSourceFieldDto = {
  name: string
  type: string
}

/**
 * DTO for table details within a data source
 */
export type DataSourceTableDto = {
  name: string
  fields: DataSourceFieldDto[]
}

/**
 * DTO for detailed data source information
 */
export type DataSourceDto = {
  id: string
  externalName: string
  type: string
  tables?: DataSourceTableDto[]
}
