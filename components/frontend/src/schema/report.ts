import { z, ZodIssueCode } from 'zod'

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
    database: z.string().refine((val) => val && val.length > 0, {
      params: { id: 'report_database_required' }
    }),
    table: z.string().refine((val) => val && val.length > 0, {
      params: { id: 'report_table_required' }
    }),
    field: z.string().refine((val) => val && val.length > 0, {
      params: { id: 'report_field_required' }
    }),
    operator: z.enum(VALID_OPERATORS),
    values: z.union([z.string(), z.array(z.string())])
  })
  .refine(
    (data) => {
      if (
        !data.values ||
        (Array.isArray(data.values) && data.values.length === 0)
      ) {
        return false
      }

      if (data.operator === 'between' && Array.isArray(data.values)) {
        return (
          data.values.length === 2 && data.values.every((v) => v && v.trim())
        )
      }

      if (Array.isArray(data.values)) {
        return data.values.some((v) => v && v.trim())
      }

      return typeof data.values === 'string' && data.values.trim().length > 0
    },
    {
      params: { id: 'report_values_invalid' }
    }
  )

export const templateId = z.string().refine((val) => val && val.length > 0, {
  params: { id: 'report_template_id_required' }
})
export const fields = z.array(filterFieldSchema)

export const report = {
  templateId,
  fields
}
