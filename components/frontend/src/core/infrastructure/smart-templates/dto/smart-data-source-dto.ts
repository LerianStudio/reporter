export type SmartDataSourceField = string

export type SmartDataSourceTable = {
  name: string
  fields: SmartDataSourceField[]
}

export type SmartDataSourceDto = {
  id: string
  externalName: string
  type: string
  tables?: SmartDataSourceTable[]
}
