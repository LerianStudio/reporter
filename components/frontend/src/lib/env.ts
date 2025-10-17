import { getRuntimeEnv } from '@lerianstudio/console-layout'

/**
 * Validates if a string is a valid URL
 */
const isValidURL = (urlString: string): boolean => {
  try {
    const url = new URL(urlString)
    // Only allow http and https protocols for security
    return url.protocol === 'http:' || url.protocol === 'https:'
  } catch {
    return false
  }
}

/**
 * Validates and returns a URL with fallback
 */
const getValidatedURL = (
  primary: string | undefined,
  secondary: string | undefined,
  fallback: string
): string => {
  if (primary && isValidURL(primary)) {
    return primary
  }
  if (secondary && isValidURL(secondary)) {
    return secondary
  }
  if (!isValidURL(fallback)) {
    throw new Error(`Invalid fallback URL: ${fallback}`)
  }
  return fallback
}

/**
 * Environment configuration utility
 * Provides type-safe access to environment variables with URL validation
 */
export const env = {
  /**
   * Documentation URL for Reporter
   */
  DOCS_URL: getValidatedURL(
    getRuntimeEnv('NEXT_PUBLIC_DOCS_URL'),
    process.env.NEXT_PUBLIC_DOCS_URL,
    'https://docs.lerian.studio/en/reporter'
  ),

  /**
   * Reporter UI Base Path (can be relative or absolute URL)
   */
  REPORTER_UI_BASE_PATH: (() => {
    const primary = getRuntimeEnv('NEXT_PUBLIC_REPORTER_UI_BASE_PATH')
    const secondary = process.env.NEXT_PUBLIC_REPORTER_UI_BASE_PATH

    // Allow relative paths for base paths
    if (primary && (primary.startsWith('/') || isValidURL(primary))) {
      return primary
    }
    if (secondary && (secondary.startsWith('/') || isValidURL(secondary))) {
      return secondary
    }
    return undefined
  })(),

  /**
   * Midaz Console Base URL
   */
  MIDAZ_CONSOLE_BASE_URL: (() => {
    const primary = getRuntimeEnv('NEXT_PUBLIC_MIDAZ_CONSOLE_BASE_URL')
    const secondary = process.env.NEXT_PUBLIC_MIDAZ_CONSOLE_BASE_URL

    if (primary && isValidURL(primary)) {
      return primary
    }
    if (secondary && isValidURL(secondary)) {
      return secondary
    }
    return undefined
  })()
}
