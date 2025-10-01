import { Page, Locator, expect } from '@playwright/test'
import { REPORT_SELECTORS } from '../fixtures/report.fixture'
import { BASE_URL } from '../fixtures/config'
import { selectOption, inputType } from './form'
import { click } from './element'

export interface ReportSearchResult {
  actionButton: Locator | null
  reportRow: Locator | null
  pageNumber: number
  found: boolean
}

export interface ReportFilterFormValues {
  database?: string | number
  table?: string | number
  field?: string | number
  operator?: string | number
  values?: string
}

/**
 * Navigates to the Reports section of the application.
 * This function handles the initial navigation and automatically expands filters if needed.
 *
 * @param page - Playwright page object
 * @returns Promise<void>
 */
export async function navigateReports(page: Page): Promise<void> {
  await page.goto(BASE_URL, { timeout: 120000 })
  await page.waitForLoadState('domcontentloaded')

  const reportsTab = page.getByRole('tab', { name: /reports/i })
  await reportsTab.click()
  await page.waitForTimeout(3000)

  await expandFilters(page)
}

/**
 * Finds a report by searching and returns its action button.
 * This function uses the reports search input to filter results and gets the first match.
 *
 * @param page - Playwright page object
 * @param searchTerm - Term to search for (could be report ID, template name, etc.)
 * @param clearSearch - Whether to clear the search input first (default: true)
 * @returns Promise<ReportSearchResult> - Object containing the action button, report row, and found status
 */
export async function findReport(
  page: Page,
  searchTerm: string,
  clearSearch: boolean = true
): Promise<ReportSearchResult> {
  let actionButton: Locator | null = null
  let reportRow: Locator | null = null
  let found = false

  try {
    // Check if reports table exists
    const tableExists = await page
      .getByTestId(REPORT_SELECTORS.table)
      .isVisible()
      .catch(() => false)

    if (!tableExists) {
      return {
        actionButton: null,
        reportRow: null,
        pageNumber: 0,
        found: false
      }
    }

    // Try to expand filters section if it exists and is collapsed
    await expandFilters(page)

    // Fill the search input (inputType handles visibility checks internally)
    await inputType(page, REPORT_SELECTORS.searchInput, searchTerm, {
      clear: clearSearch
    })

    // Wait for search results to load
    await page.waitForTimeout(1500)

    // Look for the report in the filtered results (should be the first/only result)
    const reportLocator = page.getByText(searchTerm, { exact: false }).first()
    const reportExists = await reportLocator.isVisible().catch(() => false)

    if (reportExists) {
      const allActionButtons = page.locator(
        `[data-testid^="${REPORT_SELECTORS.actionButton('').split('-').slice(0, -1).join('-')}-"]`
      )
      const actionButtonCount = await allActionButtons.count()

      if (actionButtonCount > 0) {
        const button = allActionButtons.first()
        const parentRow = button.locator('xpath=ancestor::tr[1]')

        const hasReport = await parentRow
          .getByText(searchTerm, { exact: false })
          .isVisible()
          .catch(() => false)

        if (hasReport) {
          actionButton = button
          reportRow = parentRow
          found = true
        }
      }
    }

    return {
      actionButton,
      reportRow,
      pageNumber: 1, // Search results are always on "page 1"
      found
    }
  } catch (error) {
    console.warn('Error in findReport:', error)
    return {
      actionButton: null,
      reportRow: null,
      pageNumber: 0,
      found: false
    }
  }
}

/**
 * Finds a report by ID and returns its action button.
 *
 * @param page - Playwright page object
 * @param reportId - ID of the report to find
 * @returns Promise<ReportSearchResult> - Object containing the action button, report row, and found status
 */
export async function findReportById(
  page: Page,
  reportId: string
): Promise<ReportSearchResult> {
  return findReport(page, reportId, true)
}

/**
 * Clears the search input to show all reports.
 *
 * @param page - Playwright page object
 * @returns Promise<boolean> - True if clearing was successful, false otherwise
 */
export async function clearSearch(page: Page): Promise<boolean> {
  try {
    // Try to expand filters section if it exists and is collapsed
    await expandFilters(page)

    // inputType handles visibility checks internally and returns false if element not found
    const searchSuccess = await inputType(
      page,
      REPORT_SELECTORS.searchInput,
      '',
      { clear: true }
    )

    if (!searchSuccess) {
      return false
    }
    await page.waitForTimeout(500) // Wait for results to refresh
    return true
  } catch (error) {
    console.warn('Error clearing report search:', error)
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
    const expandButton = page.getByTestId(REPORT_SELECTORS.filtersExpandButton)

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
      await click(page, REPORT_SELECTORS.filtersExpandButton, {
        waitAfterClick: 1000
      })
      return true
    }

    return true
  } catch (error) {
    console.warn('Error expanding filters section:', error)
    // Don't fail the operation, just continue
    return true
  }
}

/**
 * Switches between grid and table view modes.
 *
 * @param page - Playwright page object
 * @param mode - The desired view mode ('grid' | 'table')
 * @returns Promise<boolean> - True if successful, false otherwise
 */
export async function switchViewMode(
  page: Page,
  mode: 'grid' | 'table'
): Promise<boolean> {
  try {
    const viewToggle = page.getByTestId(REPORT_SELECTORS.viewModeToggle)
    const toggleExists = await viewToggle.isVisible().catch(() => false)

    if (!toggleExists) {
      return false
    }

    // Get current mode by checking which icon is visible
    const isGridMode = await page
      .locator('[data-testid="' + REPORT_SELECTORS.viewModeToggle + '"] svg')
      .first()
      .getAttribute('data-lucide')
      .then((attr) => attr === 'layout-grid')
      .catch(() => false)

    const currentMode = isGridMode ? 'grid' : 'table'

    // Only click if we need to switch modes
    if (currentMode !== mode) {
      await click(page, REPORT_SELECTORS.viewModeToggle, {
        waitAfterClick: 500
      })
    }

    return true
  } catch (error) {
    console.warn('Error switching view mode:', error)
    return false
  }
}

/**
 * Downloads a report by clicking the download action.
 *
 * @param page - Playwright page object
 * @param reportId - ID of the report to download
 * @returns Promise<boolean> - True if download was initiated successfully, false otherwise
 */
export async function downloadReport(
  page: Page,
  reportId: string
): Promise<boolean> {
  try {
    const reportSearch = await findReportById(page, reportId)

    if (!reportSearch.found || !reportSearch.actionButton) {
      return false
    }

    // Click the action button to open the dropdown
    await reportSearch.actionButton.click()
    await page.waitForTimeout(500)

    // Find the download option
    const downloadOption = page.getByTestId(
      REPORT_SELECTORS.downloadOption(reportId)
    )

    const downloadExists = await downloadOption.isVisible().catch(() => false)

    if (!downloadExists) {
      return false
    }

    // Click download
    await downloadOption.click()
    await page.waitForTimeout(1000)

    return true
  } catch (error) {
    console.warn('Error downloading report:', error)
    return false
  }
}

/**
 * Fills out multiple report filters using the addReportFilter method.
 * This function accepts an array of filter data objects for adding multiple filters.
 *
 * @param page - Playwright page object
 * @param filtersData - Array of filter objects, each containing values for database, table, field, operator, and values
 * @returns Promise<boolean> - Whether all filters were added successfully
 *
 * @example
 * // Add a single filter
 * await fillReportFilter(page, [
 *   { database: 'postgres_db', table: 'users', field: 'email', operator: 'equals', values: 'test@example.com' }
 * ])
 *
 * @example
 * // Add multiple filters
 * await fillReportFilter(page, [
 *   { database: 'postgres_db', table: 'users', field: 'status', operator: 'equals', values: 'active' },
 *   { database: 'postgres_db', table: 'users', field: 'created_at', operator: 'greater_than', values: '2024-01-01' }
 * ])
 */
export async function fillReportFilter(
  page: Page,
  filtersData: ReportFilterFormValues[] = []
): Promise<boolean> {
  try {
    // If no filters provided, add a default filter
    if (filtersData.length === 0) {
      filtersData = [{ values: 'test-filter-value' }]
    }

    // Add each filter using the addReportFilter method
    for (const filterData of filtersData) {
      const success = await addReportFilter(page, filterData)
      if (!success) {
        console.warn('Failed to add filter:', filterData)
        return false
      }
    }

    return true
  } catch (error) {
    console.warn('Error filling report filters:', error)
    await page.keyboard.press('Escape').catch(() => {})
    return false
  }
}

/**
 * Adds a single filter to the reports filter form by clicking the Add button and filling the fields.
 * This function allows adding multiple filters by calling it multiple times.
 *
 * @param page - Playwright page object
 * @param filterData - Object containing values for the filter fields
 * @param filterData.database - Database selection (text, number, or position)
 * @param filterData.table - Table selection (text, number, or position)
 * @param filterData.field - Field selection (text, number, or position)
 * @param filterData.operator - Operator selection (text, number, or position)
 * @param filterData.values - Values to enter in the input field
 * @returns Promise<boolean> - Whether the filter was successfully added
 *
 * @example
 * // Add a basic filter
 * await addReportFilter(page, {
 *   database: 'postgres_db',
 *   table: 'users',
 *   field: 'email',
 *   operator: 'equals',
 *   values: 'test@example.com'
 * })
 *
 * @example
 * // Add filter using indices
 * await addReportFilter(page, {
 *   database: 0,      // First database
 *   table: 'first',   // First table
 *   field: 2,         // Third field
 *   operator: 'last', // Last operator
 *   values: 'search_term'
 * })
 *
 * @example
 * // Add multiple filters
 * await addReportFilter(page, { database: 'db1', table: 'users', field: 'status', operator: 'equals', values: 'active' })
 * await addReportFilter(page, { database: 'db1', table: 'users', field: 'created_at', operator: 'greater_than', values: '2024-01-01' })
 */
export async function addReportFilter(
  page: Page,
  filterData: ReportFilterFormValues
): Promise<boolean> {
  try {
    const {
      database = 'first',
      table = 'first',
      field = 'first',
      operator = 'first',
      values = ''
    } = filterData

    // Click the Add Filter button
    // click helper handles visibility checks internally
    const addButtonSuccess = await click(
      page,
      REPORT_SELECTORS.form.addFilterButton,
      { waitAfterClick: 500 }
    )

    if (!addButtonSuccess) {
      console.warn('Add filter button not found or not clickable')
      return false
    }

    // Get all filter items to find the newly added one (should be the last one)
    const filterItems = page.locator(
      `[data-testid="${REPORT_SELECTORS.form.filterItem}"]`
    )
    const filterCount = await filterItems.count()

    if (filterCount === 0) {
      console.warn('No filter items found after clicking add button')
      return false
    }

    // Work with the last (newest) filter item to get its ID
    const lastFilterIndex = filterCount - 1
    const newFilter = filterItems.nth(lastFilterIndex)

    // Verify the new filter is visible
    await expect(newFilter).toBeVisible()

    // Work with the newly added filter using locator-based approach since testIds are dynamic
    // Find the select elements within the new filter container
    const databaseSelect = newFilter
      .locator('select, [role="combobox"]')
      .first()
    const tableSelect = newFilter.locator('select, [role="combobox"]').nth(1)
    const fieldSelect = newFilter.locator('select, [role="combobox"]').nth(2)
    const operatorSelect = newFilter.locator('select, [role="combobox"]').nth(3)
    const valuesInput = newFilter.locator('input[type="text"], textarea').last()

    // Verify all form elements are visible for this specific filter
    await expect(databaseSelect).toBeVisible()
    await expect(tableSelect).toBeVisible()
    await expect(fieldSelect).toBeVisible()
    await expect(operatorSelect).toBeVisible()

    // Use the specialized filterSelectOption helper for dynamic filter elements
    // 1. Database Selection (no waiting for enabled, it should be available initially)
    const databaseSuccess = await filterSelectOption(
      page,
      databaseSelect,
      database,
      {
        waitForEnabled: false
      }
    )
    if (!databaseSuccess) return false

    // 2. Table Selection (wait for enabled after database selection)
    const tableSuccess = await filterSelectOption(page, tableSelect, table, {
      waitForEnabled: true
    })
    if (!tableSuccess) return false

    // 3. Field Selection (wait for enabled after table selection)
    const fieldSuccess = await filterSelectOption(page, fieldSelect, field, {
      waitForEnabled: true
    })
    if (!fieldSuccess) return false

    // 4. Operator Selection (wait for enabled after field selection)
    const operatorSuccess = await filterSelectOption(
      page,
      operatorSelect,
      operator,
      {
        waitForEnabled: true
      }
    )
    if (!operatorSuccess) return false

    // 5. Values Input (if provided)
    if (values) {
      const valuesInputExists = await valuesInput.isVisible().catch(() => false)
      if (valuesInputExists) {
        await valuesInput.clear()
        await valuesInput.fill(values)
        await page.waitForTimeout(300)

        // Verify the value was entered
        await expect(valuesInput).toHaveValue(values)
      }
    }

    return true
  } catch (error) {
    console.warn('Error adding report filter:', error)
    await page.keyboard.press('Escape').catch(() => {})
    return false
  }
}

/**
 * Specialized select option method for dynamic filter elements.
 * This function handles the process of clicking a select element, waiting for options to load,
 * and selecting an option by text content, index, or position.
 *
 * @param page - Playwright page object
 * @param selectElement - The Locator for the select element (combobox, select, etc.)
 * @param value - Text content, index, or position of the option to select ('first', 'last', number, or text content)
 * @param options - Configuration options for the select operation
 * @param options.waitForEnabled - Whether to wait for the field to be enabled before clicking (default: true)
 * @param options.timeout - Timeout for waiting operations in milliseconds (default: 5000)
 * @param options.optionsTimeout - Timeout for waiting for options to appear (default: 3000)
 * @param options.waitAfterSelection - Time to wait after selection in milliseconds (default: 1000)
 * @param options.fallbackToFirst - Whether to fallback to first option if text not found (default: true)
 * @returns Promise<boolean> - Whether the option was successfully selected
 *
 * @example
 * // Basic usage with a locator
 * const databaseSelect = newFilter.locator('[role="combobox"]').first()
 * await filterSelectOption(page, databaseSelect, 'postgres_db')
 *
 * @example
 * // Select by position
 * await filterSelectOption(page, tableSelect, 'first')
 * await filterSelectOption(page, fieldSelect, 2)
 *
 * @example
 * // Select with custom options
 * await filterSelectOption(page, operatorSelect, 'equals', {
 *   waitForEnabled: false,
 *   waitAfterSelection: 500,
 *   fallbackToFirst: false
 * })
 */
export async function filterSelectOption(
  page: Page,
  selectElement: Locator,
  value: string | number = 'first',
  options: {
    waitForEnabled?: boolean
    timeout?: number
    optionsTimeout?: number
    waitAfterSelection?: number
    fallbackToFirst?: boolean
  } = {}
): Promise<boolean> {
  try {
    const {
      waitForEnabled = true,
      timeout = 5000,
      optionsTimeout = 3000,
      waitAfterSelection = 1000,
      fallbackToFirst = true
    } = options

    // Check if element exists and is visible
    const elementExists = await selectElement.isVisible().catch(() => false)
    if (!elementExists) {
      console.warn('Select element not found or not visible')
      return false
    }

    // Wait for element to be enabled if requested
    if (waitForEnabled) {
      const isEnabled = await selectElement
        .isEnabled({ timeout })
        .catch(() => false)
      if (!isEnabled) {
        console.warn('Select element is not enabled')
        return false
      }
    }

    // Click the select element to open options
    await selectElement.click()
    await page.waitForTimeout(300)

    // Wait for options to be populated
    await page.waitForFunction(
      () => {
        const options = document.querySelectorAll('[role="option"]')
        return options.length > 0
      },
      { timeout: optionsTimeout }
    )

    const optionsLocator = page.locator('[role="option"]')
    const optionCount = await optionsLocator.count()

    if (optionCount === 0) {
      await page.keyboard.press('Escape')
      return false
    }

    // Select the option based on the parameter
    let selectedOption = null

    if (value === 'first') {
      selectedOption = optionsLocator.first()
    } else if (value === 'last') {
      selectedOption = optionsLocator.last()
    } else if (typeof value === 'number') {
      if (value >= 0 && value < optionCount) {
        selectedOption = optionsLocator.nth(value)
      } else {
        console.warn(
          `Option index ${value} is out of range (0-${optionCount - 1})`
        )
        await page.keyboard.press('Escape')
        return false
      }
    } else if (typeof value === 'string') {
      // Search for option by text content
      selectedOption = optionsLocator.filter({ hasText: value }).first()
      const hasText = await selectedOption.isVisible().catch(() => false)
      if (!hasText) {
        if (fallbackToFirst) {
          // Fallback to first option if text not found
          console.warn(
            `Option with text "${value}" not found, selecting first option`
          )
          selectedOption = optionsLocator.first()
        } else {
          console.warn(
            `Option with text "${value}" not found, no fallback enabled`
          )
          await page.keyboard.press('Escape')
          return false
        }
      }
    }

    if (selectedOption) {
      await selectedOption.click()
      await page.waitForTimeout(waitAfterSelection) // Wait for selection to complete and potential API calls
      return true
    } else {
      await page.keyboard.press('Escape')
      return false
    }
  } catch (error) {
    console.warn('Error selecting option in filter element:', error)
    await page.keyboard.press('Escape').catch(() => {})
    return false
  }
}
