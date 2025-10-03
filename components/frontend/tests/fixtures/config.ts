import dotenvx from '@dotenvx/dotenvx'

// Load environment variables with expansion
dotenvx.config({ path: './.env.playwright', overload: true })

// Export BASE_URL with fallback
export const BASE_URL =
  process.env.BASE_URL || 'http://localhost/plugin-reporter-ui'

export const BASE_API_URL = process.env.BASE_API_URL || 'http://localhost:4005'

export const ORGANIZATION_ID = process.env.ORGANIZATION_ID ?? ''
