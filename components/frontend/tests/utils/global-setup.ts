import { chromium } from '@playwright/test'
import { BASE_URL } from '../fixtures/config'

async function globalSetup() {
  if (!BASE_URL) {
    throw new Error(
      'BASE_URL is not defined. Please check your .env.playwright file.'
    )
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
