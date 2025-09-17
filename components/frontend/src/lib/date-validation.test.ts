import {
  validateDateString,
  formatDateForDisplay,
  formatDateForInput
} from './date-validation'

describe('date-validation', () => {
  describe('validateDateString', () => {
    it('should return valid for undefined input', () => {
      const result = validateDateString(undefined)
      expect(result.isValid).toBe(true)
      expect(result.date).toBeUndefined()
    })

    it('should return valid for null input', () => {
      const result = validateDateString(null)
      expect(result.isValid).toBe(true)
      expect(result.date).toBeUndefined()
    })

    it('should validate correct date format', () => {
      const result = validateDateString('2024-01-15')
      expect(result.isValid).toBe(true)
      expect(result.date).toEqual(new Date(2024, 0, 15))
    })

    it('should reject invalid date format', () => {
      const result = validateDateString('2024/01/15')
      expect(result.isValid).toBe(false)
      expect(result.error?.type).toBe('format')
      expect(result.error?.message).toBe(
        'Invalid date format. Expected YYYY-MM-DD format'
      )
    })

    it('should reject invalid dates', () => {
      const result = validateDateString('2024-02-30')
      expect(result.isValid).toBe(false)
      expect(result.error?.type).toBe('invalid')
    })

    it('should reject dates outside reasonable range', () => {
      const result = validateDateString('1800-01-01')
      expect(result.isValid).toBe(false)
      expect(result.error?.type).toBe('range')
    })

    it('should accept custom date range options', () => {
      const result = validateDateString('1800-01-01', {
        minYear: 1700,
        maxYear: 1900
      })
      expect(result.isValid).toBe(true)
      expect(result.date?.getFullYear()).toBe(1800)
    })

    it('should handle parsing errors gracefully', () => {
      const result = validateDateString('2024-ab-cd')
      expect(result.isValid).toBe(false)
      expect(result.error?.type).toBe('format')
    })
  })

  describe('formatDateForDisplay', () => {
    it('should format date for display', () => {
      const date = new Date(2024, 0, 15) // January 15, 2024
      const formatted = formatDateForDisplay(date)
      expect(formatted).toBe('Jan 15, 2024')
    })
  })

  describe('formatDateForInput', () => {
    it('should format date for input', () => {
      const date = new Date(2024, 0, 15) // January 15, 2024
      const formatted = formatDateForInput(date)
      expect(formatted).toBe('2024-01-15')
    })
  })
})
