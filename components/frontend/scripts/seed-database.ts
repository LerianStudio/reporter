#!/usr/bin/env tsx

import dotenvx from '@dotenvx/dotenvx'

// Load environment variables with expansion
dotenvx.config({ path: './.env.playwright', overload: true })

import {
  cleanupTestDatabase,
  setupTestDatabase
} from '../tests/setup/setup-test-database'

async function main() {
  const args = process.argv.slice(2)
  const command = args[0] || 'setup'

  console.log(`🚀 Running test database ${command}...`)

  try {
    switch (command) {
      case 'setup':
        await setupTestDatabase()
        break

      case 'verify':
        // const isValid = await verifyTestDatabase()
        // process.exit(isValid ? 0 : 1)
        break

      case 'clean':
        await cleanupTestDatabase()
        break

      case 'reset':
        console.log('🔄 Resetting test database...')
        await cleanupTestDatabase()
        await setupTestDatabase()
        break

      default:
        console.error(`❌ Unknown command: ${command}`)
        console.log('Available commands:')
        console.log('  setup    - Setup test database (default)')
        console.log('  verify   - Verify test database setup')
        console.log('  clean    - Clean up test database')
        console.log('  reset    - Reset test database')
        process.exit(1)
    }

    console.log('✅ Operation completed successfully')
  } catch (error) {
    console.error('❌ Operation failed:', error)
    process.exit(1)
  }
}

// Run main function directly
main().catch(console.error)
