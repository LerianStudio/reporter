import { test, expect } from '@playwright/test'
import { BASE_URL } from '../fixtures/config'

test('should render application', async ({ page }) => {
  await page.goto(BASE_URL)

  await expect(page.getByTestId('message')).toBeVisible()
})
