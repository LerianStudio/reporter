import { z } from 'zod'

// Filter field schema for individual filter criteria
export const filterFieldSchema = z.object({
  database: z.string().min(1),
  table: z.string().min(1),
  field: z.string().min(1),
  values: z.string().min(1)
})

export const templateId = z.string().min(1)
export const fields = z.array(filterFieldSchema)

export const report = {
  templateId,
  fields
}
