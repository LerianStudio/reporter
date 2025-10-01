import { Page, Locator } from '@playwright/test'
import { TEMPLATE_SELECTORS } from '../fixtures/template.fixture'
import { BASE_URL } from '../fixtures/config'

export interface TemplateSearchResult {
  actionButton: Locator | null
  templateRow: Locator | null
  pageNumber: number
  found: boolean
}

/**
 * Navigates to the Templates section of the application.
 * This function handles the initial navigation and automatically expands filters if needed.
 *
 * @param page - Playwright page object
 * @returns Promise<void>
 */
export async function navigateTemplates(page: Page): Promise<void> {
  await page.goto(BASE_URL, { timeout: 120000 })
  await page.waitForLoadState('domcontentloaded')

  const templatesTab = page.getByRole('tab', { name: /templates/i })
  await templatesTab.click()
  await page.waitForTimeout(3000)

  await expandFilters(page)
}

/**
 * Finds a template by name using the search filter and returns its action button.
 * This function uses the templates search input to filter results and gets the first match.
 *
 * @param page - Playwright page object
 * @param templateName - Name of the template to find
 * @param clearSearch - Whether to clear the search input first (default: true)
 * @returns Promise<TemplateSearchResult> - Object containing the action button, template row, and found status
 */
export async function findTemplate(
  page: Page,
  templateName: string,
  clearSearch: boolean = true
): Promise<TemplateSearchResult> {
  let actionButton: Locator | null = null
  let templateRow: Locator | null = null
  let found = false

  try {
    // Check if templates table exists
    const tableExists = await page
      .getByTestId(TEMPLATE_SELECTORS.table)
      .isVisible()
      .catch(() => false)

    if (!tableExists) {
      return {
        actionButton: null,
        templateRow: null,
        pageNumber: 0,
        found: false
      }
    }

    // Try to expand filters section if it exists and is collapsed
    await expandFilters(page)

    // Get the search input
    const searchInput = page.getByTestId(TEMPLATE_SELECTORS.searchInput)
    await searchInput.isVisible().catch(() => false)

    // Clear search input if requested
    if (clearSearch) {
      await searchInput.clear()
    }

    // Type the template name in the search input
    await searchInput.fill(templateName)

    // Wait for search results to load
    await page.waitForTimeout(1500)

    // Look for the template in the filtered results (should be the first/only result)
    const templateNameLocator = page
      .getByText(templateName, { exact: true })
      .first()
    const templateExists = await templateNameLocator
      .isVisible()
      .catch(() => false)

    if (templateExists) {
      const allActionButtons = page.locator(
        `[data-testid^="${TEMPLATE_SELECTORS.actionButton('').split('-').slice(0, -1).join('-')}-"]`
      )
      const actionButtonCount = await allActionButtons.count()

      if (actionButtonCount > 0) {
        const button = allActionButtons.first()
        const parentRow = button.locator('xpath=ancestor::tr[1]')

        const hasTemplate = await parentRow
          .getByText(templateName, { exact: true })
          .isVisible()
          .catch(() => false)

        if (hasTemplate) {
          actionButton = button
          templateRow = parentRow
          found = true
        }
      }
    }

    return {
      actionButton,
      templateRow,
      pageNumber: 1, // Search results are always on "page 1"
      found
    }
  } catch (error) {
    console.warn('Error in findTemplateActionButton:', error)
    return {
      actionButton: null,
      templateRow: null,
      pageNumber: 0,
      found: false
    }
  }
}

/**
 * Clears the search input to show all templates.
 *
 * @param page - Playwright page object
 * @returns Promise<boolean> - True if clearing was successful, false otherwise
 */
export async function clearSearch(page: Page): Promise<boolean> {
  try {
    // Try to expand filters section if it exists and is collapsed
    await expandFilters(page)

    const searchInput = page.getByTestId(TEMPLATE_SELECTORS.searchInput)
    const searchInputExists = await searchInput.isVisible().catch(() => false)

    if (!searchInputExists) {
      return false
    }

    await searchInput.clear()
    await page.waitForTimeout(500) // Wait for results to refresh
    return true
  } catch (error) {
    console.warn('Error clearing template search:', error)
    return false
  }
}

/**
 * Expands the filters section if it exists and is collapsed.
 * This function handles the collapsible filters UI component.
 *
 * @param page - Playwright page object
 * @returns Promise<boolean> - True if expansion was successful or already expanded, false if filters don't exist
 */
export async function expandFilters(page: Page): Promise<boolean> {
  try {
    // Look for the expand/collapse button for filters using the test ID
    const expandButton = page.getByTestId(
      TEMPLATE_SELECTORS.filtersExpandButton
    )

    const buttonExists = await expandButton.isVisible().catch(() => false)

    if (!buttonExists) {
      return true
    }

    const isExpandedAttr = await expandButton
      .getAttribute('aria-expanded')
      .catch(() => null)
    const dataStateAttr = await expandButton
      .getAttribute('data-state')
      .catch(() => null)

    if (isExpandedAttr === 'false' || dataStateAttr === 'closed') {
      await expandButton.click()
      await page.waitForTimeout(1000)
      return true
    }

    return true
  } catch (error) {
    console.warn('Error expanding filters section:', error)
    // Don't fail the operation, just continue
    return true
  }
}
