import { Module } from '@lerianstudio/sindarian-server'
import { TemplateRepository } from '@/core/domain/repositories/template-repository'
import { SmartTemplateRepository } from '../smart-templates/repositories/smart-template-repository'
import { SmartTemplatesHttpService } from '../smart-templates/services/smart-templates-http-service'
import { ReportRepository } from '@/core/domain/repositories/report-repository'
import { SmartReportRepository } from '../smart-templates/repositories/smart-report-repository'
import { DataSourceRepository } from '@/core/domain/repositories/data-source-repository'
import { SmartDataSourceRepository } from '../smart-templates/repositories/smart-data-source-repository'

@Module({
  providers: [
    SmartTemplatesHttpService,
    {
      provide: DataSourceRepository,
      useClass: SmartDataSourceRepository
    },
    {
      provide: TemplateRepository,
      useClass: SmartTemplateRepository
    },
    {
      provide: ReportRepository,
      useClass: SmartReportRepository
    }
  ]
})
export class SmartTemplatesModule {}
