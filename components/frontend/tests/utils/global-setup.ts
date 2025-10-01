import { chromium } from '@playwright/test'
import { BASE_URL } from '../fixtures/config'
import { setupTestDatabase } from '../setup/setup-test-database'

async function globalSetup() {
  if (!BASE_URL) {
    throw new Error(
      'BASE_URL is not defined. Please check your .env.playwright file.'
    )
  }

  // Seed the database with E2E test data using backend API
  console.log('🌱 Setting up test database for E2E tests...')
  try {
    await setupTestDatabase()
    console.log('✅ Test database setup completed')
  } catch (error: Error | any) {
    console.warn(
      '⚠️  Test database setup failed, continuing with tests:',
      error.message
    )
    // Don't fail the setup if seeding fails, as the backend might not be available
    // or the tests might work without seeding
  }

  const browser = await chromium.launch()
  const page = await browser.newPage()

  try {
    await page.goto(BASE_URL, { timeout: 30000 })
    await page.waitForLoadState('networkidle', { timeout: 30000 })
  } catch (error) {
    console.error('Global Setup - Failed to load:', BASE_URL, error)
    throw error
  } finally {
    await browser.close()
  }
}

export default globalSetup
