// Report entity definitions for E2E testing
// These reports will be seeded into the database for testing

export interface ReportEntity {
  id?: string
  templateId: string
  status: 'completed' | 'processing' | 'failed' | 'pending'
  createdAt: Date
  updatedAt: Date
  metadata?: {
    fileName?: string
    fileSize?: number
    processingTime?: number
    requestedBy?: string
  }
  filters?: any
}

export const E2E_REPORTS: ReportEntity[] = [
  {
    templateId: '550e8400-e29b-41d4-a716-446655440100',
    status: 'completed',
    createdAt: new Date('2024-01-15T11:00:00Z'),
    updatedAt: new Date('2024-01-15T11:05:00Z'),
    metadata: {
      fileName: 'financial_report_2024_01_15.pdf',
      fileSize: 2048576, // 2MB
      processingTime: 300000, // 5 minutes in ms
      requestedBy: 'e2e-test-user'
    },
    filters: {}
  },
  {
    templateId: '550e8400-e29b-41d4-a716-446655440101',
    status: 'processing',
    createdAt: new Date('2024-01-16T15:00:00Z'),
    updatedAt: new Date('2024-01-16T15:02:00Z'),
    metadata: {
      fileName: 'sales_dashboard_2024_01_16.html',
      requestedBy: 'e2e-test-user'
    },
    filters: {}
  }
]

// Helper function to get report by ID
export const getReportById = (id: string): ReportEntity | undefined => {
  return E2E_REPORTS.find((report) => report.id === id)
}

// Helper function to get reports by template ID
export const getReportsByTemplateId = (templateId: string): ReportEntity[] => {
  return E2E_REPORTS.filter((report) => report.templateId === templateId)
}

// Helper function to get reports by status
export const getReportsByStatus = (
  status: ReportEntity['status']
): ReportEntity[] => {
  return E2E_REPORTS.filter((report) => report.status === status)
}

// Report selectors for E2E testing
export const REPORT_SELECTORS = {
  newButton: 'new-report-button',
  table: 'reports-table',
  tableBody: 'reports-table-body',
  emptyState: 'reports-empty-state',
  searchInput: 'reports-search-input',
  clearFilters: 'clear-filters-button',
  filtersExpandButton: 'filters-expand-button',
  dateFilterButton: 'date-filter-button',
  viewModeToggle: 'reports-view-mode-toggle',
  reportRowPrefix: 'report-row-',
  actionButton: (id: string) => `report-actions-button-${id}`,
  actionsMenu: (id: string) => `report-actions-menu-${id}`,
  downloadOption: (id: string) => `report-download-${id}`,
  tableHeaders: {
    name: 'reports-table-header-name',
    reportId: 'reports-table-header-reportid',
    status: 'reports-table-header-status',
    format: 'reports-table-header-format',
    completedAt: 'reports-table-header-completedat',
    actions: 'reports-table-header-actions'
  },
  gridView: {
    container: 'reports-grid-container',
    card: (id: string) => `report-card-${id}`,
    cardName: (id: string) => `report-card-name-${id}`,
    cardStatus: (id: string) => `report-card-status-${id}`,
    cardActions: (id: string) => `report-card-actions-${id}`
  },
  form: {
    sheet: 'reports-sheet',
    templateSelect: 'report-template-select',
    generateButton: 'report-generate-button',
    detailsTab: 'report-details-tab',
    filtersTab: 'report-filters-tab',
    addFilterButton: 'report-add-filter-button',
    filterItem: 'report-filter-item',
    filterRemoveButton: 'report-filter-remove-button',
    filterDatabaseSelect: 'report-filter-database-select',
    filterTableSelect: 'report-filter-table-select',
    filterFieldSelect: 'report-filter-field-select',
    filterOperatorSelect: 'report-filter-operator-select',
    filterValuesInput: 'report-filter-values-input'
  },
  dialog: {
    container: 'dialog',
    confirmButton: 'confirm',
    cancelButton: 'cancel'
  }
}

// Helper functions for report selectors
export const REPORT_SELECTOR_HELPERS = {
  // Get test data by ID
  getReportById: (id: string) => E2E_REPORTS.find((report) => report.id === id),

  // Selector builders
  getReportActionSelector: (reportId: string) =>
    REPORT_SELECTORS.actionButton(reportId),

  getReportMenuSelector: (reportId: string) =>
    REPORT_SELECTORS.actionsMenu(reportId),

  getReportDownloadSelector: (reportId: string) =>
    REPORT_SELECTORS.downloadOption(reportId),

  // Report row selector helpers
  getReportRowSelector: (reportId: string) =>
    `${REPORT_SELECTORS.reportRowPrefix}${reportId}`,

  getReportRowsSelector: () =>
    `[data-testid^="${REPORT_SELECTORS.reportRowPrefix}"]`,

  // Dialog selector helpers
  getDialogSelector: () => REPORT_SELECTORS.dialog.container,
  getConfirmButtonSelector: () => REPORT_SELECTORS.dialog.confirmButton,
  getCancelButtonSelector: () => REPORT_SELECTORS.dialog.cancelButton
}

export default E2E_REPORTS
