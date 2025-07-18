import { chromium } from '@playwright/test'
import { BASE_URL } from '../fixtures/config'

async function globalSetup() {
  const browser = await chromium.launch()
  const page = await browser.newPage()

  await page.goto(BASE_URL)
  await page.waitForLoadState('networkidle')
}

export default globalSetup
