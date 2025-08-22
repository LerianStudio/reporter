import { Module } from '@lerianstudio/sindarian-server'
import { DataSourceController } from '@/core/application/controllers/data-source-controller'
import { ReportController } from '@/core/application/controllers/report-controller'
import { TemplateController } from '@/core/application/controllers/template-controller'

@Module({
  controllers: [DataSourceController, TemplateController, ReportController]
})
export class ControllerModule {}
