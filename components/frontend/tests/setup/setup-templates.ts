import * as fs from 'fs'
import * as path from 'path'
import { TEMPLATES } from '../fixtures/template.fixture'
import { postFormDataRequest } from '../utils/fetcher'
import { BASE_API_URL, ORGANIZATION_ID } from '../fixtures/config'

/**
 * Create a template via the backend API
 */
async function createTemplate(template: any) {
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
  formData.append('organizationId', ORGANIZATION_ID)
  formData.append('description', template.description)
  formData.append('outputFormat', template.outputFormat)
  formData.append('fileName', template.fileName)

  if (template.metadata) {
    formData.append('metadata', JSON.stringify(template.metadata))
  }

  try {
    const response = await postFormDataRequest(
      `${BASE_API_URL}/v1/templates`,
      formData,
      {
        headers: {
          'X-Organization-Id': ORGANIZATION_ID!
        }
      }
    )
    return response
  } catch (error) {
    console.error(`Failed to create template ${template.fileName}:`, error)
    throw error
  }
}

/**
 * Setup test templates in the database via API
 * Creates templates from the E2E_TEMPLATES fixture
 */
export async function setupTemplates() {
  if (!ORGANIZATION_ID) {
    throw new Error('ORGANIZATION_ID environment variable is required')
  }

  try {
    console.log('Creating templates...')

    const createdTemplates: Record<string, any> = {}

    // Create all templates from the fixtures
    for (const template of TEMPLATES) {
      const result = await createTemplate(template)
      createdTemplates[template.fileName] = result
      console.log(`✓ Created template: ${template.description}`)
    }

    console.log('✓ Test templates created successfully')

    return createdTemplates
  } catch (error) {
    console.error('Failed to setup templates:', error)
    throw error
  }
}
