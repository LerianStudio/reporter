import { z } from 'zod'

// List of valid operators for UI validation only
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

export const filterFieldSchema = z
  .object({
    database: z.string().min(1),
    table: z.string().min(1),
    field: z.string().min(1),
    operator: z.enum(VALID_OPERATORS),
    values: z.union([z.string(), z.array(z.string())])
  })
  .refine(
    (data) => {
      // Basic UI validation for better user experience
      // The backend will handle the actual business logic validation

      if (
        !data.values ||
        (Array.isArray(data.values) && data.values.length === 0)
      ) {
        return false
      }

      // Optional: Provide helpful UI hints for specific operators
      if (data.operator === 'between' && Array.isArray(data.values)) {
        return (
          data.values.length === 2 && data.values.every((v) => v && v.trim())
        )
      }

      // For array values, ensure at least one non-empty value
      if (Array.isArray(data.values)) {
        return data.values.some((v) => v && v.trim())
      }

      // For string values, ensure it's not empty
      return typeof data.values === 'string' && data.values.trim().length > 0
    },
    {
      params: { id: 'report_values_invalid' }
    }
  )

export const templateId = z.string().min(1)
export const fields = z.array(filterFieldSchema)

export const report = {
  templateId,
  fields
}
