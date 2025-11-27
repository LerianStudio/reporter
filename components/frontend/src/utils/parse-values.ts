/**
 * Parse values to array, handling both array and comma-separated string inputs.
 * Trims whitespace and filters empty values.
 *
 * @param values - Either a string (potentially comma-separated) or an array of strings
 * @returns An array of trimmed, non-empty string values
 *
 * @example
 * // String input
 * parseValuesToArray('a, b, c') // ['a', 'b', 'c']
 *
 * @example
 * // Array input
 * parseValuesToArray(['  a  ', 'b']) // ['a', 'b']
 */
export function parseValuesToArray(values: string | string[]): string[] {
  if (Array.isArray(values)) {
    return values.map((value) => value.trim()).filter((value) => value !== '')
  }
  return values
    .split(',')
    .map((value) => value.trim())
    .filter((value) => value !== '')
}
