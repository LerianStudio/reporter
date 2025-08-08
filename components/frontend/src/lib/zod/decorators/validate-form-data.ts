import { ZodSchema } from 'zod'
import { ValidationApiException } from '@/lib/http'

export function ValidateFormData(schema: ZodSchema): MethodDecorator {
  return function (
    target: Object,
    propertyKey: string | symbol | undefined,
    descriptor: PropertyDescriptor
  ) {
    const originalMethod = descriptor.value

    descriptor.value = async function (request: Request, ...args: any[]) {
      const formData = await request.clone().formData()

      // Convert FormData to object for validation
      const data: Record<string, any> = {}

      for (const [key, value] of formData.entries()) {
        if (value instanceof File) {
          data[key] = value
        } else {
          data[key] = value
        }
      }

      // Validate the form data against the provided schema
      const parsed = schema.safeParse(data)
      if (!parsed.success) {
        // If validation fails, throw an error with details
        const firstError = parsed.error.issues[0]
        throw new ValidationApiException(
          `${firstError.path.join('.')}: ${firstError.message}`
        )
      }

      return await originalMethod.apply(this, [request, ...args])
    }
  }
}
