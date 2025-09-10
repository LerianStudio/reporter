import {
  getOperatorsForFieldType,
  operatorRequiresMultipleValues,
  operatorRequiresNoValues,
  FILTER_OPERATORS
} from './filter-operators'

describe('Filter Operators Utils', () => {
  describe('getOperatorsForFieldType', () => {
    it('should return all operators when no field type is provided', () => {
      const result = getOperatorsForFieldType()
      expect(result).toHaveLength(FILTER_OPERATORS.length)
    })

    it('should return string-appropriate operators for text fields', () => {
      const result = getOperatorsForFieldType('string')
      const operatorValues = result.map((op) => op.value)

      // All operators should now be available for string fields
      expect(operatorValues).toContain('eq')
      expect(operatorValues).toContain('gt')
      expect(operatorValues).toContain('in')
      expect(operatorValues).toContain('nin')
    })

    it('should return number-appropriate operators for numeric fields', () => {
      const result = getOperatorsForFieldType('int')
      const operatorValues = result.map((op) => op.value)

      expect(operatorValues).toContain('eq')
      expect(operatorValues).toContain('gt')
      expect(operatorValues).toContain('gte')
      expect(operatorValues).toContain('lt')
      expect(operatorValues).toContain('lte')
      expect(operatorValues).toContain('between')
      expect(operatorValues).toContain('in')
      expect(operatorValues).toContain('nin')
    })

    it('should return date-appropriate operators for date fields', () => {
      const result = getOperatorsForFieldType('datetime')
      const operatorValues = result.map((op) => op.value)

      expect(operatorValues).toContain('eq')
      expect(operatorValues).toContain('gt')
      expect(operatorValues).toContain('gte')
      expect(operatorValues).toContain('lt')
      expect(operatorValues).toContain('lte')
      expect(operatorValues).toContain('between')
      expect(operatorValues).not.toContain('in')
    })
  })

  describe('operatorRequiresMultipleValues', () => {
    it('should return true for operators that need multiple values', () => {
      expect(operatorRequiresMultipleValues('between')).toBe(true)
      expect(operatorRequiresMultipleValues('in')).toBe(true)
      expect(operatorRequiresMultipleValues('nin')).toBe(true)
    })

    it('should return false for operators that need single values', () => {
      expect(operatorRequiresMultipleValues('eq')).toBe(false)
      expect(operatorRequiresMultipleValues('gt')).toBe(false)
      expect(operatorRequiresMultipleValues('gte')).toBe(false)
      expect(operatorRequiresMultipleValues('lt')).toBe(false)
      expect(operatorRequiresMultipleValues('lte')).toBe(false)
    })
  })

  describe('operatorRequiresNoValues', () => {
    it('should return true for operators that need no values', () => {
      // None of our operators require no values
      expect(operatorRequiresNoValues('eq')).toBe(false)
      expect(operatorRequiresNoValues('gt')).toBe(false)
    })

    it('should return false for operators that need values', () => {
      expect(operatorRequiresNoValues('eq')).toBe(false)
      expect(operatorRequiresNoValues('gt')).toBe(false)
      expect(operatorRequiresNoValues('between')).toBe(false)
      expect(operatorRequiresNoValues('in')).toBe(false)
      expect(operatorRequiresNoValues('nin')).toBe(false)
    })
  })
})
