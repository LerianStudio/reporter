// Template entity definitions for E2E testing
// These templates will be seeded into the database for testing

export type TemplateEntity = {
  id?: string
  organizationId: string
  description: string
  outputFormat: string
  fileName: string
  createdAt: Date
  updatedAt: Date
  metadata?: {
    version?: string
    author?: string
    tags?: string[]
  }
}

export const E2E_TEMPLATES: TemplateEntity[] = [
  {
    organizationId: '019885e0-c544-74d4-b87c-83f89bd1be30',
    description: 'E2E Test Financial Report Template',
    outputFormat: 'TXT',
    fileName: 'receipt.tpl',
    createdAt: new Date('2024-01-15T10:00:00Z'),
    updatedAt: new Date('2024-01-15T10:00:00Z'),
    metadata: {
      version: '1.0.0',
      author: 'E2E Test Suite',
      tags: ['financial', 'report', 'pdf']
    }
  },
  {
    organizationId: '019885e0-c544-74d4-b87c-83f89bd1be30',
    description: 'E2E Test Sales Dashboard Template',
    outputFormat: 'TXT',
    fileName: 'receipt.tpl',
    createdAt: new Date('2024-01-16T14:30:00Z'),
    updatedAt: new Date('2024-01-16T14:30:00Z'),
    metadata: {
      version: '1.2.0',
      author: 'E2E Test Suite',
      tags: ['sales', 'dashboard', 'html']
    }
  }
]

// Helper function to get template by ID
export const getTemplateById = (id: string): TemplateEntity | undefined => {
  return E2E_TEMPLATES.find((template) => template.id === id)
}

// Helper function to get templates by organization
export const getTemplatesByOrganization = (
  organizationId: string
): TemplateEntity[] => {
  return E2E_TEMPLATES.filter(
    (template) => template.organizationId === organizationId
  )
}

// Template selectors for E2E testing
export const TEMPLATE_SELECTORS = {
  newButton: 'new-template-button',
  table: 'templates-table',
  tableBody: 'templates-table-body',
  emptyState: 'templates-empty-state',
  searchInput: 'templates-search-input',
  clearFilters: 'clear-filters-button',
  filtersExpandButton: 'filters-expand-button',
  outputFormatFilter: 'output-format-filter',
  dateFilterButton: 'date-filter-button',
  templateSheet: 'template-sheet',
  templateRowPrefix: 'template-row-',
  actionButton: (id: string) => `template-actions-button-${id}`,
  actionsMenu: (id: string) => `template-actions-menu-${id}`,
  detailsOption: (id: string) => `template-details-${id}`,
  deleteOption: (id: string) => `template-delete-${id}`,
  tableHeaders: {
    name: 'templates-table-header-name',
    type: 'templates-table-header-type',
    modified: 'templates-table-header-modified',
    actions: 'templates-table-header-actions'
  },
  form: {
    nameInput: 'template-name-input',
    outputFormatSelect: 'template-output-format-select',
    fileInput: 'template-file-input',
    fileInputHidden: 'template-file-input-hidden',
    submitButton: 'template-submit-button'
  },
  dialog: {
    container: 'dialog',
    confirmButton: 'confirm',
    cancelButton: 'cancel'
  }
}

// Helper functions for template selectors
export const TEMPLATE_SELECTOR_HELPERS = {
  // Get test data by ID
  getTemplateById: (id: string) =>
    E2E_TEMPLATES.find((template) => template.id === id),

  // Selector builders
  getTemplateActionSelector: (templateId: string) =>
    TEMPLATE_SELECTORS.actionButton(templateId),

  getTemplateMenuSelector: (templateId: string) =>
    TEMPLATE_SELECTORS.actionsMenu(templateId),

  getTemplateDetailsSelector: (templateId: string) =>
    TEMPLATE_SELECTORS.detailsOption(templateId),

  getTemplateDeleteSelector: (templateId: string) =>
    TEMPLATE_SELECTORS.deleteOption(templateId),

  // Template row selector helpers
  getTemplateRowSelector: (templateId: string) =>
    `${TEMPLATE_SELECTORS.templateRowPrefix}${templateId}`,

  getTemplateRowsSelector: () =>
    `[data-testid^="${TEMPLATE_SELECTORS.templateRowPrefix}"]`,

  // Dialog selector helpers
  getDialogSelector: () => TEMPLATE_SELECTORS.dialog.container,
  getConfirmButtonSelector: () => TEMPLATE_SELECTORS.dialog.confirmButton,
  getCancelButtonSelector: () => TEMPLATE_SELECTORS.dialog.cancelButton
}

export default E2E_TEMPLATES
