import { Module } from '@lerianstudio/sindarian-server'
import { TemplateRepository } from '@/core/domain/repositories/template-repository'
import { ReporterTemplateRepository } from '../reporter/repositories/reporter-template-repository'
import { ReporterHttpService } from '../reporter/services/reporter-http-service'
import { ReportRepository } from '@/core/domain/repositories/report-repository'
import { ReporterReportRepository } from '../reporter/repositories/reporter-report-repository'
import { DataSourceRepository } from '@/core/domain/repositories/data-source-repository'
import { ReporterDataSourceRepository } from '../reporter/repositories/reporter-data-source-repository'

@Module({
  providers: [
    ReporterHttpService,
    {
      provide: DataSourceRepository,
      useClass: ReporterDataSourceRepository
    },
    {
      provide: TemplateRepository,
      useClass: ReporterTemplateRepository
    },
    {
      provide: ReportRepository,
      useClass: ReporterReportRepository
    }
  ]
})
export class ReporterModule {}
