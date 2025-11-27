import { parseValuesToArray } from './parse-values'

describe('parseValuesToArray', () => {
  describe('string input', () => {
    it('should return single value in array when no comma present', () => {
      expect(parseValuesToArray('single-value')).toEqual(['single-value'])
    })

    it('should split comma-separated values into array', () => {
      expect(parseValuesToArray('val1,val2,val3')).toEqual([
        'val1',
        'val2',
        'val3'
      ])
    })

    it('should return empty array for empty string', () => {
      expect(parseValuesToArray('')).toEqual([])
    })

    it('should trim leading and trailing whitespace from each value', () => {
      expect(parseValuesToArray('  val1  ,  val2  ,  val3  ')).toEqual([
        'val1',
        'val2',
        'val3'
      ])
    })

    it('should handle spaces around commas', () => {
      expect(parseValuesToArray('a , b , c')).toEqual(['a', 'b', 'c'])
    })
  })

  describe('edge cases', () => {
    it('should filter out empty values from consecutive commas', () => {
      expect(parseValuesToArray('val1,,val2,,,val3')).toEqual([
        'val1',
        'val2',
        'val3'
      ])
    })

    it('should handle leading comma', () => {
      expect(parseValuesToArray(',val1,val2')).toEqual(['val1', 'val2'])
    })

    it('should handle trailing comma', () => {
      expect(parseValuesToArray('val1,val2,')).toEqual(['val1', 'val2'])
    })

    it('should return empty array for comma-only string', () => {
      expect(parseValuesToArray(',')).toEqual([])
      expect(parseValuesToArray(',,,')).toEqual([])
    })

    it('should return empty array for whitespace-only segments', () => {
      expect(parseValuesToArray('  ,  ,  ')).toEqual([])
    })
  })

  describe('array input', () => {
    it('should return array unchanged when no trimming needed', () => {
      expect(parseValuesToArray(['val1', 'val2'])).toEqual(['val1', 'val2'])
    })

    it('should trim whitespace from array elements', () => {
      expect(parseValuesToArray(['  val1  ', '  val2  '])).toEqual([
        'val1',
        'val2'
      ])
    })

    it('should filter empty string elements', () => {
      expect(parseValuesToArray(['val1', '', 'val2', ''])).toEqual([
        'val1',
        'val2'
      ])
    })

    it('should filter whitespace-only elements', () => {
      expect(parseValuesToArray(['val1', '   ', 'val2', '  '])).toEqual([
        'val1',
        'val2'
      ])
    })

    it('should return empty array for empty input array', () => {
      expect(parseValuesToArray([])).toEqual([])
    })
  })
})
