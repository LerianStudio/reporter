import * as fs from 'fs'
import * as path from 'path'
import E2E_TEMPLATES from '../fixtures/template.fixture'
import {
  deleteRequest,
  getRequest,
  postFormDataRequest
} from '../utils/fetcher'
import { BASE_API_URL, ORGANIZATION_ID } from '../fixtures/config'

/**
 * Seeds template data using the backend API
 */
export async function setupTemplates(): Promise<void> {
  for (const template of E2E_TEMPLATES) {
    try {
      // Create FormData with template file and metadata
      const formData = new FormData()

      // Read the template file
      const templatePath = path.join(
        __dirname,
        '../fixtures/templates',
        template.fileName
      )
      const templateContent = fs.readFileSync(templatePath, 'utf-8')
      const templateBlob = new Blob([templateContent], { type: 'text/plain' })

      // Add file to form data
      formData.append('template', templateBlob, template.fileName)

      // Add metadata as JSON string or individual fields
      formData.append('organizationId', template.organizationId)
      formData.append('description', template.description)
      formData.append('outputFormat', template.outputFormat)
      formData.append('fileName', template.fileName)

      if (template.metadata) {
        formData.append('metadata', JSON.stringify(template.metadata))
      }

      const response = await postFormDataRequest(
        `${BASE_API_URL}/v1/templates`,
        formData,
        {
          headers: {
            'X-Organization-Id': template.organizationId
          }
        }
      )

      console.log(`  ✓ Template created with ID: ${response.id}`)
    } catch (error) {
      console.error(`  ❌ Failed to create template ${template.id}:`, error)
      throw error
    }
  }
}

/**
 * Cleans up test templates by searching for templates with specific naming patterns
 * and deleting them through the backend API
 */
export async function cleanupTestTemplates(): Promise<void> {
  console.log('🧹 Cleaning up test templates...')

  try {
    // Search patterns for test templates
    const searchPatterns = ['E2E Test', 'E2E-Test', 'e2e test', 'e2e-test']

    for (const pattern of searchPatterns) {
      try {
        // List templates with name query parameter
        const templates = await getRequest(
          `${BASE_API_URL}/v1/templates?name=${encodeURIComponent(pattern)}`,
          {
            headers: {
              'X-Organization-Id': ORGANIZATION_ID
            }
          }
        )

        if (templates && Array.isArray(templates.items)) {
          console.log(
            `  Found ${templates.total} templates matching "${pattern}"`
          )

          // Delete all templates in parallel
          const deletePromises = templates.items.map(async (template: any) => {
            try {
              console.log(
                `  Deleting template: ${template.id} - ${template.description}`
              )
              await deleteRequest(
                `${BASE_API_URL}/v1/templates/${template.id}`,
                {
                  headers: {
                    'X-Organization-Id': ORGANIZATION_ID
                  }
                }
              )

              console.log(`  ✓ Deleted template: ${template.id}`)
              return { success: true, templateId: template.id }
            } catch (deleteError) {
              console.warn(
                `  ⚠️ Failed to delete template ${template.id}:`,
                deleteError
              )
              return {
                success: false,
                templateId: template.id,
                error: deleteError
              }
            }
          })

          // Wait for all delete operations to complete
          const results = await Promise.allSettled(deletePromises)

          // Log summary of results
          const successful = results.filter(
            (result) => result.status === 'fulfilled' && result.value.success
          ).length
          const failed = results.length - successful

          console.log(`  ✓ Deleted ${successful} templates, ${failed} failed`)
        }
      } catch (searchError) {
        console.warn(
          `  ⚠️ Failed to search templates with pattern "${pattern}":`,
          searchError
        )
      }
    }

    console.log('✅ Test template cleanup completed')
  } catch (error) {
    console.error('❌ Test template cleanup failed:', error)
    throw error
  }
}
