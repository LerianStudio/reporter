import { E2E_REPORTS } from '../fixtures/report.fixture'
import { postRequest } from '../utils/fetcher'
import { BASE_API_URL, ORGANIZATION_ID } from '../fixtures/config'

interface CreateReportRequest {
  templateId: string
  filters?: any
}

/**
 * Seeds report data using the backend API
 */
export async function setupReports(): Promise<void> {
  for (const report of E2E_REPORTS) {
    try {
      const reportData: CreateReportRequest = {
        templateId: report.templateId,
        filters: report.filters
      }

      const response = await postRequest(
        `${BASE_API_URL}/v1/reports`,
        reportData,
        {
          headers: { 'X-Organization-ID': ORGANIZATION_ID }
        }
      )
      console.log(`  ✓ Report created with ID: ${response.id}`)
    } catch (error) {
      throw error
    }
  }
}
