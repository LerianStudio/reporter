import dotenvx from '@dotenvx/dotenvx'

// Load environment variables with expansion
dotenvx.config({ path: './.env.playwright', overload: true })

export const {
  MIDAZ_CONSOLE_HOST,
  MIDAZ_CONSOLE_PORT,
  MIDAZ_USERNAME,
  MIDAZ_PASSWORD,
  MIDAZ_BASE_PATH,
  MIDAZ_TRANSACTION_BASE_PATH,
  ORGANIZATION_ID,
  LEDGER_ID,
  DB_HOST,
  DB_PORT,
  DB_USER,
  DB_PASSWORD,
  DB_NAME
} = process.env

// Export BASE_URL with fallback
export const BASE_URL =
  process.env.BASE_URL || 'http://localhost/plugin-reporter-ui'

export const BASE_API_URL = process.env.BASE_API_URL || 'http://localhost:4005'
