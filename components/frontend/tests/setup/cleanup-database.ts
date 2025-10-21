import { resolve } from 'path'
import { command } from '../utils/cmd'
import { delay } from '../utils/delay'

// Get the correct paths to docker-compose files
const componentsInfraPath = resolve(
  __dirname,
  '../../../infra/docker-compose.test.yml'
)

const testsInfraPath = resolve(__dirname, '../infra')

/**
 * Cleanup test database and Docker environment
 * This will:
 * 1. Stop all test containers from both infra locations
 * 2. Remove all test volumes
 * 3. Optionally restart the containers
 */
export async function cleanupDatabase(restart = false) {
  try {
    console.log('Starting cleanup of test environment...\n')

    // Step 1: Stop and remove containers and volumes from components/infra
    console.log(
      '1. Stopping components/infra containers and removing volumes...'
    )
    const componentsInfraDownCommand = `docker compose -f "${componentsInfraPath}" down -v`

    try {
      await command(componentsInfraDownCommand)
      console.log('✓ Components/infra containers stopped and volumes removed')
    } catch (error) {
      console.error('Failed to stop components/infra containers:', error)
      throw error
    }

    // Step 2: Stop and remove containers and volumes from tests/infra
    console.log('\n2. Stopping tests/infra containers and removing volumes...')
    const testsInfraDownCommands = [
      `docker compose -f "${testsInfraPath}/docker-compose.infra.yml" down -v`
    ]

    try {
      for (const downCommand of testsInfraDownCommands) {
        await command(downCommand)
      }
      console.log('✓ Tests/infra containers stopped and volumes removed')
    } catch (error) {
      console.error('Failed to stop tests/infra containers:', error)
      throw error
    }

    if (restart) {
      // Step 3: Restart components/infra containers
      console.log('\n3. Restarting components/infra containers...')
      const componentsInfraUpCommand = `docker compose -f "${componentsInfraPath}" up -d`

      try {
        await command(componentsInfraUpCommand)
        console.log('✓ Components/infra containers restarted')
      } catch (error) {
        console.error('Failed to restart components/infra containers:', error)
        throw error
      }

      // Step 4: Restart tests/infra containers
      console.log('\n4. Restarting tests/infra containers...')
      const testsInfraUpCommands = [
        `docker compose -f "${testsInfraPath}/docker-compose.infra.yml" up -d`
      ]

      try {
        for (const upCommand of testsInfraUpCommands) {
          await command(upCommand)
        }
        console.log('✓ Tests/infra containers restarted')
      } catch (error) {
        console.error('Failed to restart tests/infra containers:', error)
        throw error
      }

      // Wait for containers to be healthy
      console.log('\n5. Waiting for containers to be healthy...')
      await delay(10000)
      console.log('✓ Containers should be ready')

      // Step 6: Restart specific services to run migrations
      console.log(
        '\n6. Restarting Onboarding, Transaction, and Manager services to run migrations...'
      )

      try {
        // Restart onboarding and transaction from tests/infra
        await command(
          `docker compose -f "${testsInfraPath}/docker-compose.onboarding.yml" restart`
        )
        await command(
          `docker compose -f "${testsInfraPath}/docker-compose.transaction.yml" restart`
        )

        // Restart manager from components/infra
        const workerPath = resolve(
          __dirname,
          '../../../worker/docker-compose.test.yml'
        )
        await command(`docker compose -f "${workerPath}" restart`)

        const managerPath = resolve(
          __dirname,
          '../../../manager/docker-compose.test.yml'
        )
        await command(`docker compose -f "${managerPath}" restart`)

        console.log('✓ Services restarted')
      } catch (error) {
        console.error('Failed to restart services:', error)
        throw error
      }

      // Wait for migrations to complete
      console.log('\n7. Waiting for migrations to complete...')
      await delay(5000)
      console.log('✓ Migrations should be complete')
    }

    console.log('\n✓ Cleanup completed successfully!')

    if (restart) {
      console.log('\nYou can now run: npm run db:seed')
    } else {
      console.log('\nTo start containers again:')
      console.log(
        '  - Components infra: docker compose -f components/infra/docker-compose.test.yml up -d'
      )
      console.log(
        '  - Tests infra: docker compose -f components/frontend/tests/infra/docker-compose.*.yml up -d'
      )
    }

    return true
  } catch (error) {
    console.error('\n✗ Cleanup failed:', error)
    throw error
  }
}

// Run if executed directly
if (require.main === module) {
  // Check if --restart flag is provided
  const restart = process.argv.includes('--restart')

  cleanupDatabase(restart)
    .then(() => process.exit(0))
    .catch((error) => {
      console.error(error)
      process.exit(1)
    })
}
