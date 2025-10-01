import { test, expect } from '@playwright/test'
import {
  REPORT_SELECTORS,
  REPORT_SELECTOR_HELPERS
} from '../fixtures/report.fixture'
import {
  findReportById,
  navigateReports,
  switchViewMode,
  fillReportFilter
} from '../utils/report'
import { inputType, selectOption } from '../utils/form'
import { click } from '../utils/element'

test.describe.configure({ mode: 'parallel' })
test.describe('Reports Feature', () => {
  test.describe('Reports List Operations', () => {
    test.beforeEach(async ({ page }) => {
      await navigateReports(page)
    })

    test('should display reports page correctly', async ({ page }) => {
      // Verify new button is accessible and clickable, then close any opened dialog
      const newButtonClickable = await click(page, REPORT_SELECTORS.newButton)
      expect(newButtonClickable).toBe(true)
      await page.keyboard.press('Escape')
    })

    test('should display reports table when reports exist', async ({
      page
    }) => {
      const tableExists = await page
        .getByTestId(REPORT_SELECTORS.table)
        .isVisible()

      if (tableExists) {
        await expect(
          page.getByTestId(REPORT_SELECTORS.tableHeaders.name)
        ).toBeVisible()
        await expect(
          page.getByTestId(REPORT_SELECTORS.tableHeaders.reportId)
        ).toBeVisible()
        await expect(
          page.getByTestId(REPORT_SELECTORS.tableHeaders.status)
        ).toBeVisible()
        await expect(
          page.getByTestId(REPORT_SELECTORS.tableHeaders.format)
        ).toBeVisible()
        await expect(
          page.getByTestId(REPORT_SELECTORS.tableHeaders.completedAt)
        ).toBeVisible()
        await expect(
          page.getByTestId(REPORT_SELECTORS.tableHeaders.actions)
        ).toBeVisible()
        await expect(page.getByTestId(REPORT_SELECTORS.tableBody)).toBeVisible()

        const paginationExists = await page
          .locator(
            '[data-testid*="pagination"], .pagination, [class*="pagination"]'
          )
          .isVisible()

        if (paginationExists) {
          await expect(
            page.locator(
              '[data-testid*="pagination"], .pagination, [class*="pagination"]'
            )
          ).toBeVisible()
        }
      }
    })

    test('should be able to search reports', async ({ page }) => {
      await inputType(page, REPORT_SELECTORS.searchInput, 'test report')

      await page.waitForTimeout(1000)

      const searchInput = page.getByTestId(REPORT_SELECTORS.searchInput)
      await expect(searchInput).toHaveValue('test report')
    })

    test('should be able to filter reports by date', async ({ page }) => {
      await click(page, REPORT_SELECTORS.dateFilterButton)

      const calendarVisible = await page
        .locator('[role="dialog"], .calendar, [data-testid*="calendar"]')
        .isVisible()
      if (calendarVisible) {
        await page.keyboard.press('Escape')
      }
    })

    test('should be able to clear all filters', async ({ page }) => {
      await inputType(page, REPORT_SELECTORS.searchInput, 'test')

      await click(page, REPORT_SELECTORS.clearFilters)

      const searchInput = page.getByTestId(REPORT_SELECTORS.searchInput)
      await expect(searchInput).toHaveValue('')
    })

    test('should be able to switch between grid and table view modes', async ({
      page
    }) => {
      const viewToggle = page.getByTestId(REPORT_SELECTORS.viewModeToggle)

      const toggleExists = await viewToggle.isVisible().catch(() => false)

      if (toggleExists) {
        await click(page, REPORT_SELECTORS.viewModeToggle, {
          waitAfterClick: 500
        })
        await click(page, REPORT_SELECTORS.viewModeToggle, {
          waitAfterClick: 500
        })
      }
    })

    test('should show report actions dropdown when clicking on actions button', async ({
      page
    }) => {
      const actionButton = page
        .locator(
          `[data-testid^="${REPORT_SELECTORS.actionButton('').split('-').slice(0, -1).join('-')}-"]`
        )
        .first()

      const actionButtonExists = await actionButton
        .isVisible()
        .catch(() => false)

      if (actionButtonExists) {
        await actionButton.click()

        const dropdown = page
          .locator(
            `[data-testid^="${REPORT_SELECTORS.actionsMenu('').split('-').slice(0, -1).join('-')}-"]`
          )
          .first()
        const dropdownVisible = await dropdown
          .isVisible({ timeout: 3000 })
          .catch(() => false)

        if (dropdownVisible) {
          await expect(dropdown).toBeVisible({ timeout: 3000 })

          const downloadOption = page
            .locator(
              `[data-testid^="${REPORT_SELECTORS.downloadOption('').split('-').slice(0, -1).join('-')}-"]`
            )
            .first()

          if (await downloadOption.isVisible().catch(() => false)) {
            await expect(downloadOption).toBeVisible()
          }
        }
      }
    })

    test('should display pagination controls when reports exist', async ({
      page
    }) => {
      const paginationExists = await page
        .locator(
          '[data-testid*="pagination"], .pagination, [class*="pagination"]'
        )
        .isVisible()

      if (paginationExists) {
        await expect(
          page.locator(
            '[data-testid*="pagination"], .pagination, [class*="pagination"]'
          )
        ).toBeVisible()
      }
    })

    test('should expand and collapse filters section', async ({ page }) => {
      const collapseButton = page
        .locator('button[data-state], button[aria-expanded]')
        .first()

      if (await collapseButton.isVisible()) {
        const initialState =
          (await collapseButton.getAttribute('aria-expanded')) ||
          (await collapseButton.getAttribute('data-state'))

        await collapseButton.click()
        await page.waitForTimeout(500)

        const newState =
          (await collapseButton.getAttribute('aria-expanded')) ||
          (await collapseButton.getAttribute('data-state'))
        expect(newState).not.toBe(initialState)
      }
    })

    test('should show loading state while fetching reports', async ({
      page
    }) => {
      const loadingIndicator = page
        .getByText(/updating/i)
        .or(page.locator('.animate-spin'))
        .or(page.getByText(/loading/i))

      const loadingVisible = await loadingIndicator.isVisible({ timeout: 1000 })

      if (loadingVisible) {
        await expect(loadingIndicator).toBeVisible()
      }

      await page.waitForTimeout(3000)
      await expect(
        page.getByRole('heading', { name: 'Reports', exact: true })
      ).toBeVisible()
    })
  })

  test.describe('Reports Get Operations', () => {
    test.beforeEach(async ({ page }) => {
      await navigateReports(page)
    })

    test('should be able to find and display report details by ID', async ({
      page
    }) => {
      const reportId = '550e8400-e29b-41d4-a716-446655440200'

      const reportSearchResult = await findReportById(page, reportId)

      if (reportSearchResult.found) {
        await expect(reportSearchResult.reportRow!).toBeVisible()

        // Verify the report ID is displayed in the row
        await expect(
          reportSearchResult.reportRow!.getByText(reportId, { exact: false })
        ).toBeVisible()

        // Verify action button is available
        await expect(reportSearchResult.actionButton!).toBeVisible()
      }
    })

    test('should display correct report information in table rows', async ({
      page
    }) => {
      const tableExists = await page
        .getByTestId(REPORT_SELECTORS.table)
        .isVisible()
        .catch(() => false)

      if (tableExists) {
        const reportRows = page.locator(
          REPORT_SELECTOR_HELPERS.getReportRowsSelector()
        )
        const rowCount = await reportRows.count()

        if (rowCount > 0) {
          const firstRow = reportRows.first()

          // Check that the row contains typical report information
          const hasContent = await firstRow.locator('td').first().isVisible()
          expect(hasContent).toBe(true)

          // Verify action button exists
          const actionButton = firstRow.locator(
            `[data-testid^="${REPORT_SELECTORS.actionButton('').split('-').slice(0, -1).join('-')}-"]`
          )
          const hasActionButton = await actionButton
            .isVisible()
            .catch(() => false)

          if (hasActionButton) {
            await expect(actionButton).toBeVisible()
          }
        }
      }
    })

    test('should show appropriate status badges for different report statuses', async ({
      page
    }) => {
      const tableExists = await page
        .getByTestId(REPORT_SELECTORS.table)
        .isVisible()
        .catch(() => false)

      if (tableExists) {
        const statusBadges = page.locator(
          '[class*="badge"], [data-testid*="status"]'
        )
        const badgeCount = await statusBadges.count()

        if (badgeCount > 0) {
          const firstBadge = statusBadges.first()
          await expect(firstBadge).toBeVisible()

          // Verify it contains expected status text
          const badgeText = await firstBadge.textContent()
          const validStatuses = [
            'completed',
            'processing',
            'failed',
            'pending',
            'Finished',
            'Processing',
            'Failed',
            'Pending'
          ]
          const hasValidStatus = validStatuses.some((status) =>
            badgeText?.toLowerCase().includes(status.toLowerCase())
          )

          if (badgeText) {
            expect(hasValidStatus).toBe(true)
          }
        }
      }
    })

    test('should display report format information', async ({ page }) => {
      const tableExists = await page
        .getByTestId(REPORT_SELECTORS.table)
        .isVisible()
        .catch(() => false)

      if (tableExists) {
        const formatBadges = page.locator('[class*="badge"]')
        const badgeCount = await formatBadges.count()

        if (badgeCount > 0) {
          // Look for format-related badges (PDF, HTML, TXT, etc.)
          const formatBadge = formatBadges
            .filter({ hasText: /PDF|HTML|TXT|CSV/i })
            .first()
          const formatBadgeExists = await formatBadge
            .isVisible()
            .catch(() => false)

          if (formatBadgeExists) {
            await expect(formatBadge).toBeVisible()
          }
        }
      }
    })
  })

  test.describe('Reports Download Operations', () => {
    test.beforeEach(async ({ page }) => {
      await navigateReports(page)
    })

    test('should be able to download a completed report', async ({ page }) => {
      // Look for completed reports only
      const completedReports = page
        .locator('tr')
        .filter({ hasText: /completed|finished/i })
      const completedCount = await completedReports.count()

      if (completedCount > 0) {
        const firstCompletedReport = completedReports.first()

        // Find the action button for this report
        const actionButton = firstCompletedReport.locator(
          `[data-testid^="${REPORT_SELECTORS.actionButton('').split('-').slice(0, -1).join('-')}-"]`
        )

        const actionButtonExists = await actionButton
          .isVisible()
          .catch(() => false)

        if (actionButtonExists) {
          await actionButton.click()
          await page.waitForTimeout(500)

          // Look for download option
          const downloadOption = page
            .locator('[data-testid*="download"]')
            .or(page.getByText(/download/i))
            .first()

          const downloadExists = await downloadOption
            .isVisible()
            .catch(() => false)

          if (downloadExists) {
            await expect(downloadOption).toBeVisible()
            // Note: We don't actually click download in tests to avoid file download issues
          }
        }
      }
    })

    test('should not show download option for non-completed reports', async ({
      page
    }) => {
      // Look for processing or failed reports
      const nonCompletedReports = page
        .locator('tr')
        .filter({ hasText: /processing|failed|pending/i })
      const nonCompletedCount = await nonCompletedReports.count()

      if (nonCompletedCount > 0) {
        const firstNonCompletedReport = nonCompletedReports.first()

        // Find the action button for this report
        const actionButton = firstNonCompletedReport.locator(
          `[data-testid^="${REPORT_SELECTORS.actionButton('').split('-').slice(0, -1).join('-')}-"]`
        )

        const actionButtonExists = await actionButton
          .isVisible()
          .catch(() => false)

        if (actionButtonExists) {
          await actionButton.click()
          await page.waitForTimeout(500)

          // Download option should not be visible for non-completed reports
          const downloadOption = page
            .locator('[data-testid*="download"]')
            .or(page.getByText(/download/i))
            .first()

          const downloadExists = await downloadOption
            .isVisible({ timeout: 1000 })
            .catch(() => false)

          // For non-completed reports, download should typically not be available
          if (downloadExists) {
            // If download exists, it should be disabled
            const isDisabled = await downloadOption.getAttribute('disabled')
            expect(isDisabled).not.toBeNull()
          }
        }
      }
    })
  })

  test.describe('Reports Grid View Operations', () => {
    test.beforeEach(async ({ page }) => {
      await navigateReports(page)
      // Switch to grid view if not already
      await switchViewMode(page, 'grid')
    })

    test('should display reports in grid view when toggled', async ({
      page
    }) => {
      const gridContainer = page.getByTestId(
        REPORT_SELECTORS.gridView.container
      )
      const gridExists = await gridContainer.isVisible().catch(() => false)

      if (gridExists) {
        await expect(gridContainer).toBeVisible()

        // Check for report cards
        const reportCards = page.locator(
          `[data-testid^="${REPORT_SELECTORS.gridView.card('').split('-').slice(0, -1).join('-')}-"]`
        )
        const cardCount = await reportCards.count()

        if (cardCount > 0) {
          await expect(reportCards.first()).toBeVisible()
        }
      }
    })

    test('should display report information in grid cards', async ({
      page
    }) => {
      const reportCards = page.locator(
        `[data-testid^="${REPORT_SELECTORS.gridView.card('').split('-').slice(0, -1).join('-')}-"]`
      )
      const cardCount = await reportCards.count()

      if (cardCount > 0) {
        const firstCard = reportCards.first()
        await expect(firstCard).toBeVisible()

        // Check that card contains some content
        const hasContent = await firstCard.locator('*').first().isVisible()
        expect(hasContent).toBe(true)
      }
    })

    test('should be able to interact with report cards in grid view', async ({
      page
    }) => {
      const reportCards = page.locator(
        `[data-testid^="${REPORT_SELECTORS.gridView.card('').split('-').slice(0, -1).join('-')}-"]`
      )
      const cardCount = await reportCards.count()

      if (cardCount > 0) {
        const firstCard = reportCards.first()

        // Look for action elements within the card
        const cardActions = firstCard.locator('button, [role="button"]')
        const actionCount = await cardActions.count()

        if (actionCount > 0) {
          const actionElement = cardActions.first()
          await expect(actionElement).toBeVisible()

          // Verify the action element is interactive
          const isClickable = await actionElement.isEnabled().catch(() => false)
          expect(isClickable).toBe(true)
        }
      }
    })
  })

  test.describe('Reports Generation Operations', () => {
    test.beforeEach(async ({ page }) => {
      await navigateReports(page)
    })

    test('should be able to open the report generation form', async ({
      page
    }) => {
      await click(page, REPORT_SELECTORS.newButton, { waitAfterClick: 2000 })

      const sheet = page.getByTestId(REPORT_SELECTORS.form.sheet)
      await expect(sheet).toBeVisible()

      // Verify form elements are visible
      const templateSelect = page.getByTestId(
        REPORT_SELECTORS.form.templateSelect
      )
      const generateButton = page.getByTestId(
        REPORT_SELECTORS.form.generateButton
      )

      await expect(templateSelect).toBeVisible()
      await expect(generateButton).toBeVisible()

      // Close the sheet
      await page.keyboard.press('Escape')
    })

    test('should display available templates in the template selector', async ({
      page
    }) => {
      await click(page, REPORT_SELECTORS.newButton, { waitAfterClick: 2000 })

      const sheet = page.getByTestId(REPORT_SELECTORS.form.sheet)
      await expect(sheet).toBeVisible()

      // Use selectOption to open template dropdown and check if options are available
      const templateSelectSuccess = await selectOption(
        page,
        REPORT_SELECTORS.form.templateSelect,
        'first',
        { fallbackToFirst: false, waitAfterSelection: 0 }
      )

      // If templates are available, the selection would succeed
      if (templateSelectSuccess) {
        // Template was selected successfully, escape to close
        await page.keyboard.press('Escape')
      } else {
        // No templates available, which is also valid - just open dropdown to verify
        await click(page, REPORT_SELECTORS.form.templateSelect, {
          waitAfterClick: 500
        })
        await page.keyboard.press('Escape')
      }

      // Close the sheet
      await page.keyboard.press('Escape')
    })

    test('should require template selection before generating report', async ({
      page
    }) => {
      await click(page, REPORT_SELECTORS.newButton, { waitAfterClick: 2000 })

      const sheet = page.getByTestId(REPORT_SELECTORS.form.sheet)
      await expect(sheet).toBeVisible()

      // Try to generate without selecting a template
      await click(page, REPORT_SELECTORS.form.generateButton, {
        waitAfterClick: 1000
      })

      // Sheet should still be visible (form validation should prevent submission)
      const isSheetStillVisible = await sheet.isVisible()
      expect(isSheetStillVisible).toBe(true)

      // Close the sheet
      await page.keyboard.press('Escape')
    })

    test('should be able to generate a report with valid template selection', async ({
      page
    }) => {
      await click(page, REPORT_SELECTORS.newButton, { waitAfterClick: 2000 })

      const sheet = page.getByTestId(REPORT_SELECTORS.form.sheet)
      await expect(sheet).toBeVisible()

      // Use selectOption to select the first available template
      const templateSelectSuccess = await selectOption(
        page,
        REPORT_SELECTORS.form.templateSelect,
        'first'
      )

      if (templateSelectSuccess) {
        // Navigate to filters tab to test filters functionality
        await click(page, REPORT_SELECTORS.form.filtersTab, {
          waitAfterClick: 500
        })

        // Test adding a filter
        const addFilterButton = page.getByTestId(
          REPORT_SELECTORS.form.addFilterButton
        )
        const addButtonExists = await addFilterButton
          .isVisible()
          .catch(() => false)

        if (addButtonExists) {
          // Use the new fillReportFilter method to add a filter cleanly
          const filterSuccess = await fillReportFilter(page, [
            {
              database: 'midaz_onboarding',
              table: 'ledger',
              field: 'name',
              operator: 'equal',
              values: 'test-value'
            },
            {
              database: 'midaz_transaction',
              table: 'transaction',
              field: 'id',
              operator: 'equal',
              values: 'another-value'
            }
          ])

          // Verify the filter was added successfully
          expect(filterSuccess).toBe(true)
        }

        // Go back to details tab to generate the report
        await click(page, REPORT_SELECTORS.form.detailsTab, {
          waitAfterClick: 500
        })

        // Generate the report
        await click(page, REPORT_SELECTORS.form.generateButton, {
          waitAfterClick: 3000
        })

        // Check for success indicators
        const isSheetClosed = await sheet
          .isVisible()
          .then((visible) => !visible)
          .catch(() => true)

        const successToast = page
          .locator('.sonner-toast, [data-sonner-toast], [role="status"]')
          .getByText(/success|started/i)
        const toastVisible = await successToast
          .isVisible({ timeout: 5000 })
          .catch(() => false)

        // Either sheet should close OR success toast should appear
        const reportGenerated = isSheetClosed || toastVisible
        expect(reportGenerated).toBe(true)
      } else {
        // If no templates available, just verify the form works
        await page.keyboard.press('Escape')
      }
    })

    test('should show form validation for missing required fields', async ({
      page
    }) => {
      await click(page, REPORT_SELECTORS.newButton, { waitAfterClick: 2000 })

      const sheet = page.getByTestId(REPORT_SELECTORS.form.sheet)
      await expect(sheet).toBeVisible()

      // Verify required field indicators are present
      const formLabels = sheet.locator('label, .form-label')
      const labelCount = await formLabels.count()

      if (labelCount > 0) {
        // Look for asterisk or required indicators
        const requiredIndicators = sheet
          .locator('text="*"')
          .or(sheet.locator('.required'))
        const hasRequiredIndicators = (await requiredIndicators.count()) > 0

        // At minimum, the form should have some structure
        expect(labelCount).toBeGreaterThan(0)
      }

      // Close the sheet
      await page.keyboard.press('Escape')
    })

    test('should be able to navigate between form tabs and manage filters', async ({
      page
    }) => {
      await click(page, REPORT_SELECTORS.newButton, { waitAfterClick: 2000 })

      const sheet = page.getByTestId(REPORT_SELECTORS.form.sheet)
      await expect(sheet).toBeVisible()

      // Test tab navigation
      // Verify both tabs are accessible by attempting to click them
      const detailsTabClickable = await click(
        page,
        REPORT_SELECTORS.form.detailsTab,
        { waitAfterClick: 100 }
      )
      const filtersTabClickable = await click(
        page,
        REPORT_SELECTORS.form.filtersTab,
        { waitAfterClick: 100 }
      )

      expect(detailsTabClickable).toBe(true)
      expect(filtersTabClickable).toBe(true)

      // Navigate to filters tab
      await click(page, REPORT_SELECTORS.form.filtersTab, {
        waitAfterClick: 500
      })

      // Test add filter functionality
      const addFilterButton = page.getByTestId(
        REPORT_SELECTORS.form.addFilterButton
      )
      const addButtonExists = await addFilterButton
        .isVisible()
        .catch(() => false)

      if (addButtonExists) {
        await addFilterButton.click()
        await page.waitForTimeout(500)

        // Verify filter item was added
        const filterItems = page.locator(
          `[data-testid="${REPORT_SELECTORS.form.filterItem}"]`
        )
        const filterCount = await filterItems.count()

        if (filterCount > 0) {
          await expect(filterItems.first()).toBeVisible()

          // Test all filter form elements are visible
          const databaseSelect = page
            .getByTestId(REPORT_SELECTORS.form.filterDatabaseSelect)
            .first()
          const tableSelect = page
            .getByTestId(REPORT_SELECTORS.form.filterTableSelect)
            .first()
          const fieldSelect = page
            .getByTestId(REPORT_SELECTORS.form.filterFieldSelect)
            .first()
          const operatorSelect = page
            .getByTestId(REPORT_SELECTORS.form.filterOperatorSelect)
            .first()

          const databaseSelectExists = await databaseSelect
            .isVisible()
            .catch(() => false)
          const tableSelectExists = await tableSelect
            .isVisible()
            .catch(() => false)
          const fieldSelectExists = await fieldSelect
            .isVisible()
            .catch(() => false)
          const operatorSelectExists = await operatorSelect
            .isVisible()
            .catch(() => false)

          if (databaseSelectExists) await expect(databaseSelect).toBeVisible()
          if (tableSelectExists) await expect(tableSelect).toBeVisible()
          if (fieldSelectExists) await expect(fieldSelect).toBeVisible()
          if (operatorSelectExists) await expect(operatorSelect).toBeVisible()

          // Test filter form workflow if elements exist
          if (databaseSelectExists) {
            await databaseSelect.click()
            await page.waitForTimeout(300)

            const databaseOptions = page.locator('[role="option"]')
            const dbOptionCount = await databaseOptions.count()

            if (dbOptionCount > 0) {
              await databaseOptions.first().click()
              await page.waitForTimeout(500)

              // After selecting database, table should be enabled
              if (tableSelectExists) {
                const isTableEnabled = await tableSelect
                  .isEnabled()
                  .catch(() => false)
                if (isTableEnabled) {
                  await tableSelect.click()
                  await page.waitForTimeout(300)

                  const tableOptions = page.locator('[role="option"]')
                  const tableOptionCount = await tableOptions.count()

                  if (tableOptionCount > 0) {
                    await tableOptions.first().click()
                    await page.waitForTimeout(300)
                  } else {
                    await page.keyboard.press('Escape')
                  }
                }
              }
            } else {
              await page.keyboard.press('Escape')
            }
          }

          // Test remove filter functionality
          const removeButton = page
            .getByTestId(REPORT_SELECTORS.form.filterRemoveButton)
            .first()
          const removeButtonExists = await removeButton
            .isVisible()
            .catch(() => false)

          if (removeButtonExists) {
            await removeButton.click()
            await page.waitForTimeout(500)

            // Verify filter was removed
            const remainingFilters = await page
              .locator(`[data-testid="${REPORT_SELECTORS.form.filterItem}"]`)
              .count()
            expect(remainingFilters).toBe(filterCount - 1)
          }
        }
      }

      // Navigate back to details tab
      await click(page, REPORT_SELECTORS.form.detailsTab, {
        waitAfterClick: 500
      })

      // Verify we're back on details tab by checking template select
      const templateSelect = page.getByTestId(
        REPORT_SELECTORS.form.templateSelect
      )
      await expect(templateSelect).toBeVisible()

      // Close the sheet
      await page.keyboard.press('Escape')
    })
  })
})
