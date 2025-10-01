import { z } from 'zod'
import {
  operatorRequiresNoValues,
  operatorRequiresMultipleValues
} from '@/utils/filter-operators'

const VALID_OPERATORS = [
  'eq',
  'gt',
  'gte',
  'lt',
  'lte',
  'between',
  'in',
  'nin'
] as const

const isFilterEmpty = (data: any) => {
  return (
    (!data.database || data.database.trim() === '') &&
    (!data.table || data.table.trim() === '') &&
    (!data.field || data.field.trim() === '') &&
    (!data.values ||
      (Array.isArray(data.values) && data.values.length === 0) ||
      (typeof data.values === 'string' && data.values.trim() === ''))
  )
}

const validateValues = (values: string | string[], operator: string) => {
  if (operatorRequiresNoValues(operator)) {
    return true
  }

  if (!values) {
    return { valid: false, message: 'report_values_required' }
  }

  if (operator === 'between') {
    if (!Array.isArray(values)) {
      return { valid: false, message: 'report_values_between_required' }
    }
    if (values.length !== 2 || !values.every((v) => v && v.trim())) {
      return { valid: false, message: 'report_values_between_required' }
    }
    return { valid: true }
  }

  if (operatorRequiresMultipleValues(operator)) {
    if (!Array.isArray(values)) {
      return { valid: false, message: 'report_values_multiple_required' }
    }
    if (values.length === 0 || !values.some((v) => v && v.trim())) {
      return { valid: false, message: 'report_values_multiple_required' }
    }
    return { valid: true }
  }

  if (Array.isArray(values)) {
    if (values.length === 0 || !values[0]?.trim()) {
      return { valid: false, message: 'report_values_required' }
    }
  } else if (typeof values === 'string' && !values.trim()) {
    return { valid: false, message: 'report_values_required' }
  }

  return { valid: true }
}

export const filterFieldSchema = z
  .object({
    database: z.string(),
    table: z.string(),
    field: z.string(),
    operator: z.union([z.enum(VALID_OPERATORS), z.literal('')]),
    values: z.union([z.string(), z.array(z.string())])
  })
  .superRefine((data, ctx) => {
    if (isFilterEmpty(data)) {
      return
    }

    const hasAnyData =
      (data.database && data.database.trim() !== '') ||
      (data.table && data.table.trim() !== '') ||
      (data.field && data.field.trim() !== '') ||
      (data.operator && data.operator.trim() !== '') ||
      (data.values &&
        ((Array.isArray(data.values) && data.values.length > 0) ||
          (typeof data.values === 'string' && data.values.trim() !== '')))

    if (hasAnyData) {
      if (!data.database || data.database.trim() === '') {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          path: ['database'],
          params: { id: 'report_database_required' }
        })
      }

      if (!data.table || data.table.trim() === '') {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          path: ['table'],
          params: { id: 'report_table_required' }
        })
      }

      if (!data.field || data.field.trim() === '') {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          path: ['field'],
          params: { id: 'report_field_required' }
        })
      }

      if (!data.operator || data.operator.trim() === '') {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          path: ['operator'],
          params: { id: 'report_operator_required' }
        })
      }

      if (data.operator && data.operator.trim() !== '') {
        const result = validateValues(data.values, data.operator)
        if (typeof result === 'object' && 'valid' in result && !result.valid) {
          ctx.addIssue({
            code: z.ZodIssueCode.custom,
            path: ['values'],
            params: { id: result.message || 'report_values_required' }
          })
        }
      } else if (hasAnyData) {
        const result = validateValues(data.values, 'eq')
        if (typeof result === 'object' && 'valid' in result && !result.valid) {
          ctx.addIssue({
            code: z.ZodIssueCode.custom,
            path: ['values'],
            params: { id: 'report_values_required' }
          })
        }
      }
    }
  })

export const templateId = z.string().refine((val) => val && val.length > 0, {
  params: { id: 'report_template_id_required' }
})
export const fields = z.array(filterFieldSchema)

export const report = {
  templateId,
  fields
}
