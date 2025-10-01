import { cleanupTestTemplates, setupTemplates } from './setup-templates'
import { setupReports } from './setup-reports'

/**
 * Sets up the test database by seeding templates and reports through backend API endpoints
 */
export async function setupTestDatabase(): Promise<void> {
  console.log('🔧 Setting up test database...')

  try {
    // Clean up existing test data first
    await cleanupTestDatabase()

    // Seed templates
    console.log('📝 Seeding templates...')
    await setupTemplates()

    // Seed reports
    console.log('📊 Seeding reports...')
    await setupReports()

    console.log('✅ Test database setup completed successfully')
  } catch (error) {
    console.error('❌ Failed to setup test database:', error)
    throw error
  }
}

/**
 * Cleans up existing test data
 */
export async function cleanupTestDatabase(): Promise<void> {
  console.log('🧹 Cleaning up existing test data...')

  try {
    await cleanupTestTemplates()
  } catch (error) {
    console.error('❌ Failed to cleanup test database:', error)
    throw error
  }
}
