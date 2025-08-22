import { z } from 'zod'

// Output format enum for templates
const OUTPUT_FORMATS = ['csv', 'xml', 'html', 'txt'] as const

/**
 * Validation schema for template creation form
 * Based on CreateTemplateDto and Figma design requirements
 */
export const templateFormSchema = z.object({
  // Template Name - required text field
  name: z.string().min(1).max(255),

  // Output Format - required select field
  outputFormat: z.enum(OUTPUT_FORMATS),

  // Template File - required file upload
  templateFile: z
    .instanceof(File, {
      params: { id: 'template_file_required' }
    })
    .refine((file) => file.name.endsWith('.tpl'), {
      params: { id: 'template_file_invalid_extension' }
    })
    .refine(
      (file) => file.size <= 5 * 1024 * 1024, // 5MB limit
      { params: { id: 'template_file_too_large' } }
    )
    .refine((file) => file.size > 0, { params: { id: 'template_file_empty' } })
})

/**
 * Validation schema for template update form
 * All fields are optional for updates
 */
export const templateUpdateFormSchema = z.object({
  name: z.string().min(1).max(255),

  outputFormat: z.enum(OUTPUT_FORMATS).optional(),

  templateFile: z
    .instanceof(File)
    .refine((file) => file.name.endsWith('.tpl'), {
      params: { id: 'template_file_invalid_extension' }
    })
    .refine((file) => file.size <= 5 * 1024 * 1024, {
      params: { id: 'template_file_too_large' }
    })
    .refine((file) => file.size > 0, { params: { id: 'template_file_empty' } })
})

// Type inference for form data
export type TemplateFormData = z.infer<typeof templateFormSchema>
export type TemplateUpdateFormData = z.infer<typeof templateUpdateFormSchema>

// Output format options for select component
export const OUTPUT_FORMAT_OPTIONS = OUTPUT_FORMATS.map((format) => ({
  label: format,
  value: format
}))
