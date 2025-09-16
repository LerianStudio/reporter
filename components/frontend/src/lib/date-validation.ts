import { parseISO, isValid, format } from 'date-fns'

export interface DateValidationResult {
  isValid: boolean
  date?: Date
  error?: {
    type: 'format' | 'invalid' | 'range' | 'parsing'
    message: string
  }
}

export interface DateValidationOptions {
  minYear?: number
  maxYear?: number
}

/**
 * Validates a date string and returns validation result with typed error information
 */
export function validateDateString(
  dateString: string | undefined | null,
  options: DateValidationOptions = {}
): DateValidationResult {
  // Set reasonable defaults for date range
  const currentYear = new Date().getFullYear()
  const minYear = options.minYear ?? 1900
  const maxYear = options.maxYear ?? currentYear + 10

  // Handle null/undefined input
  if (!dateString) {
    return { isValid: true, date: undefined }
  }

  // Validate date string format (YYYY-MM-DD)
  if (
    typeof dateString !== 'string' ||
    !/^\d{4}-\d{2}-\d{2}$/.test(dateString)
  ) {
    return {
      isValid: false,
      error: {
        type: 'format',
        message: 'Invalid date format. Expected YYYY-MM-DD format'
      }
    }
  }

  try {
    // Parse date using date-fns for robust handling
    const date = parseISO(dateString + 'T00:00:00')

    if (!isValid(date)) {
      return {
        isValid: false,
        error: {
          type: 'invalid',
          message: 'The selected date is not valid'
        }
      }
    }

    // Validate date is within reasonable range
    const dateYear = date.getFullYear()
    if (dateYear < minYear || dateYear > maxYear) {
      return {
        isValid: false,
        error: {
          type: 'range',
          message: `Please select a date between ${minYear} and ${maxYear}`
        }
      }
    }

    return { isValid: true, date }
  } catch (error) {
    return {
      isValid: false,
      error: {
        type: 'parsing',
        message: 'An unexpected error occurred while parsing the date'
      }
    }
  }
}

/**
 * Formats a date for display in the UI
 */
export function formatDateForDisplay(date: Date): string {
  return format(date, 'MMM dd, yyyy')
}

/**
 * Formats a date for form input (YYYY-MM-DD)
 */
export function formatDateForInput(date: Date): string {
  return format(date, 'yyyy-MM-dd')
}
