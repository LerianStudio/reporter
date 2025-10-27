export type ReporterDataSourceField =
  | {
      name: string
      type: string
    }
  | string

export type ReporterDataSourceTable = {
  name: string
  fields: ReporterDataSourceField[]
}

export type ReporterDataSourceDto = {
  id: string
  externalName: string
  type: string
  tables?: ReporterDataSourceTable[]
}
