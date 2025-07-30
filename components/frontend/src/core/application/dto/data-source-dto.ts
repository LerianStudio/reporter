/**
 * DTO for data source information
 */
export type DataSourceInformationDto = {
  id: string
  externalName: string
  type: string
}

/**
 * DTO for table details within a data source
 */
export type TableDetailsDto = {
  name: string
  fields: string[]
}

/**
 * DTO for detailed data source information
 */
export type DataSourceDetailsDto = {
  id: string
  externalName: string
  type: string
  tables: TableDetailsDto[]
} 