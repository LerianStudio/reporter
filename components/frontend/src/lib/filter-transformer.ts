import {
  FilterField,
  AdvancedReportFilters
} from '@/core/domain/entities/report-entity'

export function transformFiltersToApiFormat(
  fields: FilterField[]
): AdvancedReportFilters {
  const result: AdvancedReportFilters = {}

  fields.forEach((field) => {
    const { database, table, field: fieldName, operator, values } = field

    if (!result[database]) {
      result[database] = {}
    }
    if (!result[database][table]) {
      result[database][table] = {}
    }
    if (!result[database][table][fieldName]) {
      result[database][table][fieldName] = {}
    }

    let normalizedValues: string[]
    
    if (Array.isArray(values)) {
      normalizedValues = values
        .filter((val): val is string => val != null && typeof val === 'string')
        .map(val => val.toString())
    } else if (values != null) {
      normalizedValues = [String(values)]
    } else {
      normalizedValues = []
    }

    result[database][table][fieldName][operator] = normalizedValues
  })

  return result
}

export function parseFilterValues(valuesString: string): string[] {
  return valuesString
    .split(',')
    .map((value) => value.trim())
    .filter(Boolean)
}

export function transformToApiPayload(
  templateId: string,
  fields: FilterField[]
): {
  templateId: string
  filters: AdvancedReportFilters
} {
  const processedFields = fields.map((field) => {
    let processedValues: string[]

    if (Array.isArray(field.values)) {
      processedValues = field.values
        .filter((val): val is string => Boolean(val) && typeof val === 'string')
    } else if (field.values != null) {
      if (typeof field.values === 'string') {
        processedValues = parseFilterValues(field.values)
      } else {
        processedValues = [String(field.values)]
      }
    } else {
      processedValues = []
    }

    return {
      ...field,
      values: processedValues
    }
  })

  const transformedFilters = transformFiltersToApiFormat(processedFields)

  return {
    templateId,
    filters: transformedFilters
  }
}
