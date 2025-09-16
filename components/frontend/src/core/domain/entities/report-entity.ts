import { MetadataEntity } from './metadata-entity'
import { PaginationSearchEntity } from './pagination-entity'
import { TemplateEntity } from './template-entity'

export type ReportStatus = 'Processing' | 'Finished' | 'Failed'

export type FilterField = {
  database: string
  table: string
  field: string
  operator: string
  values: string | string[]
}

export type AdvancedReportFilters = {
  [database: string]: {
    [table: string]: {
      [field: string]: {
        [operator: string]: string[]
      }
    }
  }
}

export type ReportSearchEntity = PaginationSearchEntity & {
  status?: ReportStatus
  search?: string
  templateId?: string
}

export type ReportFilters = {
  date_range?: {
    start: string
    end: string
  }
  account_types?: string[]
  minimum_balance?: number
  maximum_balance?: number
  asset_codes?: string[]
  portfolio_ids?: string[]
  search?: string
  fields?: FilterField[]
}

export type ReportEntity = {
  id?: string
  templateId: string
  organizationId: string
  status?: ReportStatus
  filters?: ReportFilters
  template?: TemplateEntity
  metadata?: MetadataEntity
  createdAt?: Date
  updatedAt?: Date
  completedAt?: Date
}
