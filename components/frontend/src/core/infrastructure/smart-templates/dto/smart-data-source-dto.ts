export type SmartDataSourceField =
  | {
      name: string
      type: string
    }
  | string

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
