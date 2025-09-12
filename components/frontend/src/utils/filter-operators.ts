export interface FilterOperator {
  value: string
  label: string
  description: string
  fieldTypes: string[]
}

const STRING_TYPES = ['string', 'text', 'varchar']
const NUMERIC_TYPES = ['number', 'int', 'integer', 'decimal', 'float']
const DATE_TYPES = ['date', 'datetime', 'timestamp']
const BOOLEAN_TYPES = ['boolean']

const ALL_TYPES = [
  ...STRING_TYPES,
  ...NUMERIC_TYPES,
  ...DATE_TYPES,
  ...BOOLEAN_TYPES
]
const COMPARABLE_TYPES = [...STRING_TYPES, ...NUMERIC_TYPES, ...DATE_TYPES]
const IN_OUT_TYPES = [...STRING_TYPES, ...NUMERIC_TYPES]

export const FILTER_OPERATORS: FilterOperator[] = [
  {
    value: 'eq',
    label: 'Equal',
    description: 'Field value equals the specified value',
    fieldTypes: ALL_TYPES
  },
  {
    value: 'gt',
    label: 'Greater Than',
    description: 'Field value is greater than the specified value',
    fieldTypes: COMPARABLE_TYPES
  },
  {
    value: 'gte',
    label: 'Greater or Equal',
    description: 'Field value is greater than or equal to the specified value',
    fieldTypes: COMPARABLE_TYPES
  },
  {
    value: 'lt',
    label: 'Less Than',
    description: 'Field value is less than the specified value',
    fieldTypes: COMPARABLE_TYPES
  },
  {
    value: 'lte',
    label: 'Less or Equal',
    description: 'Field value is less than or equal to the specified value',
    fieldTypes: COMPARABLE_TYPES
  },
  {
    value: 'between',
    label: 'Between',
    description: 'Field value is between two specified values (inclusive)',
    fieldTypes: COMPARABLE_TYPES
  },
  {
    value: 'in',
    label: 'In',
    description: 'Field value is in the specified list of values',
    fieldTypes: [
      'string',
      'number',
      'text',
      'varchar',
      'int',
      'integer',
      'decimal',
      'float'
    ]
  },
  {
    value: 'nin',
    label: 'Not In',
    description: 'Field value is not in the specified list of values',
    fieldTypes: [
      'string',
      'number',
      'text',
      'varchar',
      'int',
      'integer',
      'decimal',
      'float'
    ]
  }
]

export function getOperatorsForFieldType(fieldType?: string): FilterOperator[] {
  if (!fieldType) {
    return FILTER_OPERATORS
  }

  const normalizedFieldType = fieldType.toLowerCase()
  return FILTER_OPERATORS.filter((operator) =>
    operator.fieldTypes.some(
      (type) =>
        type === normalizedFieldType ||
        normalizedFieldType.includes(type) ||
        type.includes(normalizedFieldType)
    )
  )
}

export function operatorRequiresMultipleValues(operatorValue: string): boolean {
  return ['between', 'in', 'nin'].includes(operatorValue)
}

export function operatorRequiresNoValues(_operatorValue: string): boolean {
  return false
}
