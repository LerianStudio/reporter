import { REPORTS } from '../fixtures/report.fixture'
import { TEMPLATES } from '../fixtures/template.fixture'
import { postRequest, getRequest } from '../utils/fetcher'
import { BASE_API_URL, ORGANIZATION_ID } from '../fixtures/config'

interface CreateReportRequest {
  templateId?: string
  filters?: any
}

/**
 * Fetch a template by description to get its ID
 */
async function getTemplateByDescription(description: string): Promise<string> {
  try {
    const templates: any = await getRequest(
      `${BASE_API_URL}/v1/templates?description=${encodeURIComponent(description)}`,
      {
        headers: {
          'X-Organization-Id': ORGANIZATION_ID!
        }
      }
    )

    if (
      templates &&
      Array.isArray(templates.items) &&
      templates.items.length > 0
    ) {
      return templates.items[0].id
    }

    throw new Error(`Template not found with description: ${description}`)
  } catch (error) {
    console.error(
      `Failed to fetch template with description ${description}:`,
      error
    )
    throw error
  }
}

/**
 * Create a report via the backend API
 */
async function createReport(
  templateId: string,
  report: CreateReportRequest & { description?: string }
) {
  const reportData: CreateReportRequest = {
    templateId,
    filters: report.filters
  }

  try {
    const response = await postRequest(
      `${BASE_API_URL}/v1/reports`,
      reportData,
      {
        headers: { 'X-Organization-ID': ORGANIZATION_ID! }
      }
    )
    return response
  } catch (error) {
    console.error(`Failed to create report for template ${templateId}:`, error)
    throw error
  }
}

/**
 * Setup test reports in the database via API
 * Creates reports from the REPORTS fixture
 * Fetches template IDs by description before creating reports
 */
export async function setupReports() {
  if (!ORGANIZATION_ID) {
    throw new Error('ORGANIZATION_ID environment variable is required')
  }

  try {
    console.log('Creating reports...')

    const createdReports: Record<string, any> = {}

    // Create all reports from the fixtures
    for (let i = 0; i < REPORTS.length; i++) {
      const report = REPORTS[i]
      const templateFixture = TEMPLATES[i]

      // Fetch the template ID by description
      console.log(`  Fetching template: ${templateFixture.description}`)
      const templateId = await getTemplateByDescription(
        templateFixture.description
      )

      // Create the report with the fetched template ID
      const result = await createReport(templateId, report)
      createdReports[templateId] = result
      console.log(
        `✓ Created report for template: ${templateFixture.description}`
      )
    }

    console.log('✓ Test reports created successfully')

    return createdReports
  } catch (error) {
    console.error('Failed to setup reports:', error)
    throw error
  }
}

// Run if executed directly
if (require.main === module) {
  setupReports()
    .then(() => process.exit(0))
    .catch((error) => {
      console.error(error)
      process.exit(1)
    })
}
