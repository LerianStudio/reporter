export type DataSourceField = {
  name: string
  type: string
}

export type DataSourceTable = {
  name: string
  fields: DataSourceField[]
}

export type DataSource = {
  id: string
  externalName: string
  type: string
  tables?: DataSourceTable[]
}
