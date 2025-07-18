import { Container, ContainerModule } from '../../utils/di/container'
import { SmartTemplatesHttpService } from '../../smart-templates/services/smart-templates-http-service'
import { SmartTemplateRepository } from '../../smart-templates/repositories/smart-template-repository'
import { SmartReportRepository } from '../../smart-templates/repositories/smart-report-repository'
import { TemplateRepository } from '@/core/domain/repositories/template-repository'
import { ReportRepository } from '@/core/domain/repositories/report-repository'

/**
 * Smart Templates Module for dependency injection
 *
 * Registers unified infrastructure for Smart Templates API integration:
 * - Unified HTTP service for both templates and reports
 * - Template repository implementation
 * - Report repository implementation
 *
 * This module replaces the separate plugin-templates and plugin-reports modules
 * with a unified approach following Clean Architecture principles.
 */
export const SmartTemplatesModule = new ContainerModule(
  (container: Container) => {
    // HTTP Service registration - unified service for both templates and reports
    container
      .bind<SmartTemplatesHttpService>(SmartTemplatesHttpService)
      .toSelf()

    // Template repository registration
    container
      .bind<TemplateRepository>(TemplateRepository)
      .to(SmartTemplateRepository)

    // Report repository registration
    container.bind<ReportRepository>(ReportRepository).to(SmartReportRepository)
  }
)
