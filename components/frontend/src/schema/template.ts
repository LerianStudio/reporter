import { z } from 'zod'
import { IntlShape } from 'react-intl'

// Output format enum for templates
const OUTPUT_FORMATS = ['csv', 'xml', 'html', 'txt', 'pdf'] as const

/**
 * Creates internationalized validation schema for template creation form
 *
 * @param intl - IntlShape object for message formatting
 * @returns Zod schema with internationalized error messages
 *
 * @example
 * ```typescript
 * // In your component with useIntl:
 * import { useIntl } from 'react-intl'
 * import { createTemplateFormSchema } from '@/schema/template'
 *
 * const intl = useIntl()
 * const schema = createTemplateFormSchema(intl)
 * const form = useForm({
 *   resolver: zodResolver(schema)
 * })
 * ```
 */
export const createTemplateFormSchema = (intl: IntlShape) =>
  z.object({
    name: z.string().min(1).max(255),
    outputFormat: z.enum(OUTPUT_FORMATS),
    templateFile: z
      .instanceof(File, {
        message: intl.formatMessage({
          id: 'errors.template.file.required',
          defaultMessage: 'Template file is required'
        })
      })
      .refine((file) => file.name.endsWith('.tpl'), {
        message: intl.formatMessage({
          id: 'errors.template.file.invalid_extension',
          defaultMessage: 'File must be a .tpl template file'
        })
      })
      .refine(
        (file) => file.size <= 5 * 1024 * 1024, // 5MB limit
        {
          message: intl.formatMessage({
            id: 'errors.template.file.too_large',
            defaultMessage: 'File size must be less than 5MB'
          })
        }
      )
      .refine((file) => file.size > 0, {
        message: intl.formatMessage({
          id: 'errors.template.file.empty',
          defaultMessage: 'File cannot be empty'
        })
      })
  })

/**
 * Creates internationalized validation schema for template update form
 *
 * @param intl - IntlShape object for message formatting
 * @returns Zod schema with internationalized error messages
 *
 * @example
 * ```typescript
 * // In your component with useIntl:
 * import { useIntl } from 'react-intl'
 * import { createTemplateUpdateFormSchema } from '@/schema/template'
 *
 * const intl = useIntl()
 * const schema = createTemplateUpdateFormSchema(intl)
 * const form = useForm({
 *   resolver: zodResolver(schema)
 * })
 * ```
 */
export const createTemplateUpdateFormSchema = (intl: IntlShape) =>
  z.object({
    name: z.string().min(1).max(255),

    outputFormat: z.enum(OUTPUT_FORMATS).optional(),

    templateFile: z
      .instanceof(File)
      .refine((file) => file.name.endsWith('.tpl'), {
        message: intl.formatMessage({
          id: 'errors.template.file.invalid_extension',
          defaultMessage: 'File must be a .tpl template file'
        })
      })
      .refine((file) => file.size <= 5 * 1024 * 1024, {
        message: intl.formatMessage({
          id: 'errors.template.file.too_large',
          defaultMessage: 'File size must be less than 5MB'
        })
      })
      .refine((file) => file.size > 0, {
        message: intl.formatMessage({
          id: 'errors.template.file.empty',
          defaultMessage: 'File cannot be empty'
        })
      })
  })

/**
 * Type definitions for form data using factory return types
 *
 * These types are inferred from the schema factory functions to ensure
 * type safety while maintaining proper internationalization support.
 *
 * @example
 * ```typescript
 * // Use these types in your form components:
 * const handleSubmit = (data: TemplateFormData) => {
 *   // data is properly typed with name, outputFormat, and templateFile
 * }
 * ```
 */
export type TemplateFormData = z.infer<
  ReturnType<typeof createTemplateFormSchema>
>
export type TemplateUpdateFormData = z.infer<
  ReturnType<typeof createTemplateUpdateFormSchema>
>

/**
 * Pre-formatted output format options for select components
 *
 * @example
 * ```typescript
 * // Use in your select component:
 * {OUTPUT_FORMAT_OPTIONS.map((option) => (
 *   <SelectItem key={option.value} value={option.value}>
 *     {option.label}
 *   </SelectItem>
 * ))}
 * ```
 */
export const OUTPUT_FORMAT_OPTIONS = OUTPUT_FORMATS.map((format) => ({
  label: format,
  value: format
}))
