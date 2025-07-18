import { MetadataEntity } from './metadata-entity'
import { TemplateEntity } from './template-entity'

// Report status types following PRD specification (Processing → Finished/Failed)
export type ReportStatus = 'Processing' | 'Finished' | 'Failed'

// Filter field type for individual filter criteria
export type FilterField = {
  database: string
  table: string
  field: string
  values: string[]
}

// Report filters for data querying
export type ReportFilters = {
  ledger_ids?: string[]
  date_range?: {
    start: string // ISO 8601 date string
    end: string // ISO 8601 date string
  }
  account_types?: string[]
  minimum_balance?: number
  maximum_balance?: number
  asset_codes?: string[]
  portfolio_ids?: string[]
  search?: string
  // Array of filter criteria for data source filtering
  fields?: FilterField[]
}

// Main report entity following Clean Architecture patterns
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
