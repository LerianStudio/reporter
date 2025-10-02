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
  const currentYear = new Date().getFullYear()
  const minYear = options.minYear ?? 1900
  const maxYear = options.maxYear ?? currentYear + 10

  if (!dateString) {
    return { isValid: true, date: undefined }
  }

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

/**
 * Validates and formats date for API queries - handles both Date objects and strings
 * Consolidates date validation logic used in repositories
 */
export function validateAndFormatDateForQuery(
  date: Date | string | undefined
): string | undefined {
  if (!date) {
    return undefined
  }

  if (date instanceof Date) {
    if (isNaN(date.getTime())) {
      throw new Error('Invalid date provided')
    }
    return date.toISOString().split('T')[0]
  }

  if (typeof date === 'string') {
    const validation = validateDateString(date)
    if (!validation.isValid) {
      const errorType = validation.error?.type || 'invalid'
      const errorMessages = {
        format: 'Invalid date format. Expected YYYY-MM-DD',
        invalid: 'Invalid date string provided',
        range: 'Date is out of acceptable range',
        parsing: 'Unable to parse date string'
      }
      throw new Error(errorMessages[errorType])
    }
    return date
  }

  throw new Error('createdAt must be a Date object or a valid date string')
}
