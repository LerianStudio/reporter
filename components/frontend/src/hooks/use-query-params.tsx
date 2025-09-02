'use client'

import { useEffect, useState } from 'react'
import debounce from 'lodash/debounce'
import isEmpty from 'lodash/isEmpty'
import pick from 'lodash/pick'
import { useForm, UseFormProps } from 'react-hook-form'
import { usePagination } from './use-pagination'
import { useSearchParams } from '@/lib/search/use-search-params'

// Branded types for better type safety
type ValidatedString = string & { __brand: 'ValidatedString' }
type ValidatedQueryKey = string & { __brand: 'ValidatedQueryKey' }

interface SerializedFormValues {
  [key: ValidatedQueryKey]: string | number | boolean | undefined | null
}

export type UseQueryParams<
  SearchParams extends Record<string, any> = Record<string, any>
> = {
  initialValues?: Partial<SearchParams>
  total: number
  formProps?: Partial<UseFormProps>
  debounce?: number
}

export function useQueryParams<
  SearchParams extends Record<string, any> = Record<string, any>
>({
  initialValues = {},
  total,
  formProps,
  debounce: debounceProp = 300
}: UseQueryParams<SearchParams>): {
  form: ReturnType<typeof useForm>
  searchValues: {
    page: string
    limit: string
  } & SearchParams
  pagination: ReturnType<typeof usePagination>
} {
  const pagination = usePagination({ total })
  const { searchParams, updateSearchParams } = useSearchParams()

  const [searchValues, setSearchValues] = useState<
    {
      page: string
      limit: string
    } & SearchParams
  >({
    page: pagination.page.toString(),
    limit: pagination.limit.toString(),
    ...(initialValues as SearchParams)
  })

  const form = useForm({
    ...formProps,
    defaultValues: {
      ...initialValues,
      page: pagination.page.toString(),
      limit: pagination.limit.toString()
    }
  })

  useEffect(() => {
    const newValues = {
      ...searchValues,
      page: pagination.page.toString(),
      limit: pagination.limit.toString()
    }

    setSearchValues(newValues)

    const queryParams = pick(searchParams, Object.keys(newValues))

    if (
      !(
        isEmpty(queryParams) &&
        newValues.page === searchParams?.page &&
        newValues.limit === searchParams?.limit
      )
    ) {
      updateSearchParams(newValues)
    }
  }, [pagination.page, pagination.limit])

  const sanitizeStringValue = (input: string): ValidatedString => {
    const inputStr = String(input)

    // Block dangerous protocols and JavaScript URLs
    const dangerousPatterns = [
      /javascript:/gi,
      /data:/gi,
      /vbscript:/gi,
      /on\w+\s*=/gi // Event handlers like onclick, onload, etc.
    ]

    let sanitized = inputStr
    let previousLength
    // Apply patterns repeatedly until no more matches are found (handles nested attacks)
    do {
      previousLength = sanitized.length
      dangerousPatterns.forEach((pattern) => {
        sanitized = sanitized.replace(pattern, '')
      })
    } while (sanitized.length !== previousLength && sanitized.length > 0)

    // Comprehensive HTML entity encoding for XSS prevention
    return (
      sanitized
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#x27;')
        .replace(/\//g, '&#x2F;')
        // Remove null bytes and control characters
        .replace(/[\0-\x1F\x7F]/g, '')
        // Limit string length to prevent DoS
        .substring(0, 500)
        .trim() as ValidatedString
    )
  }

  const isValidKey = (key: string): key is ValidatedQueryKey => {
    // Validate query parameter keys to prevent injection
    return /^[a-zA-Z][a-zA-Z0-9_]*$/.test(key) && key.length <= 50
  }

  const serializeFormValues = (
    values: Record<string, any>
  ): SerializedFormValues => {
    const serialized: SerializedFormValues = {}

    Object.entries(values).forEach(([key, value]) => {
      // Validate key format
      if (!isValidKey(key)) {
        console.warn(`Invalid query parameter key: ${key}`)
        return
      }

      // Key is now typed as ValidatedQueryKey after the guard
      const validatedKey = key as ValidatedQueryKey

      if (value === null || value === undefined) {
        serialized[validatedKey] = undefined
      } else if (typeof value === 'string') {
        // Sanitize string values to prevent XSS
        const sanitizedValue = sanitizeStringValue(value)
        serialized[validatedKey] = sanitizedValue || undefined
      } else if (typeof value === 'number') {
        // Validate numeric values
        if (isFinite(value) && !isNaN(value)) {
          serialized[validatedKey] = value
        } else {
          console.warn(`Invalid numeric value for key ${key}:`, value)
          serialized[validatedKey] = undefined
        }
      } else if (typeof value === 'boolean') {
        serialized[validatedKey] = value
      } else {
        // Convert other types to strings and sanitize
        const stringValue = String(value)
        const sanitizedValue = sanitizeStringValue(stringValue)
        serialized[validatedKey] = sanitizedValue || undefined
      }
    })

    return serialized
  }

  useEffect(() => {
    const { unsubscribe } = form.watch(
      debounce((values) => {
        const serializedValues = serializeFormValues(values)
        updateSearchParams(serializedValues)
        setSearchValues(values)
      }, debounceProp)
    )

    return () => unsubscribe()
  }, [form.watch, debounceProp])

  useEffect(() => {
    if (isEmpty(searchParams)) {
      return
    }

    const allowedKeys = ['page', 'limit', ...Object.keys(initialValues || [])]
    const value = pick(searchParams, allowedKeys)

    if (isEmpty(value)) {
      return
    }

    form.reset({
      ...initialValues,
      page: pagination.page.toString(),
      limit: pagination.limit.toString(),
      ...value
    })

    pagination.setPage(Number(value.page))
    pagination.setLimit(Number(value.limit))
  }, [])

  return { form, searchValues, pagination }
}
