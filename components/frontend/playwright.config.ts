import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  testDir: './tests/e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 1,
  workers: process.env.CI ? 1 : undefined,
  reporter: 'html',
  globalSetup: require.resolve('./tests/utils/global-setup'),
  timeout: 60000, // Increase global timeout to 60 seconds
  use: {
    baseURL: 'http://localhost:8083',
    // storageState: 'tests/storage/data.json',
    trace: 'on-first-retry',
    navigationTimeout: 30000, // 30 second navigation timeout
    actionTimeout: 10000 // 10 second action timeout
  },

  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] }
    },

    {
      name: 'firefox',
      use: { ...devices['Desktop Firefox'] }
    },

    {
      name: 'webkit',
      use: { ...devices['Desktop Safari'] }
    }
  ],

  webServer: {
    command: 'npm run dev',
    port: 8083,
    reuseExistingServer: !process.env.CI
  }
})
