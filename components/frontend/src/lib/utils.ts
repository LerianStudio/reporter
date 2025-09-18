import { type ClassValue, clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export { env } from './env'
export * from './date-validation'
export { TemplateQueryKeys } from './query-keys'
