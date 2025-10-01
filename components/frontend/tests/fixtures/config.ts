import dotenvx from '@dotenvx/dotenvx'

// Load environment variables with expansion
dotenvx.config({ path: './.env.playwright', overload: true })

// Export BASE_URL with fallback
export const BASE_URL = 'http://localhost:8083/plugin-smart-templates-ui'

export const BASE_API_URL =
  process.env.PLUGIN_SMART_TEMPLATES_BASE_PATH || 'http://localhost:4005'

export const ORGANIZATION_ID =
  process.env.ORGANIZATION_ID || '019885e0-c544-74d4-b87c-83f89bd1be30'
