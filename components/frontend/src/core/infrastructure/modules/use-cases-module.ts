import { Module } from '@lerianstudio/sindarian-server'

import { CreateTemplateUseCase } from '@/core/application/use-cases/templates/create-template-use-case'
import { DeleteTemplateUseCase } from '@/core/application/use-cases/templates/delete-template-use-case'
import { GetTemplateUseCase } from '@/core/application/use-cases/templates/get-template-use-case'
import { ListTemplatesUseCase } from '@/core/application/use-cases/templates/list-templates-use-case'
import { UpdateTemplateUseCase } from '@/core/application/use-cases/templates/update-template-use-case'
import { DownloadReportUseCase } from '@/core/application/use-cases/reports/download-report-use-case'
import { GenerateReportUseCase } from '@/core/application/use-cases/reports/generate-report-use-case'
import { GetReportStatusUseCase } from '@/core/application/use-cases/reports/get-report-status-use-case'
import { ListReportsUseCase } from '@/core/application/use-cases/reports/list-reports-use-case'
import { GetDataSourceUseCase } from '@/core/application/use-cases/data-sources/get-data-source-use-case'
import { ListDataSourcesUseCase } from '@/core/application/use-cases/data-sources/list-data-sources-use-case'

@Module({
  providers: [
    CreateTemplateUseCase,
    GetTemplateUseCase,
    ListTemplatesUseCase,
    UpdateTemplateUseCase,
    DeleteTemplateUseCase,
    DownloadReportUseCase,
    GenerateReportUseCase,
    GetReportStatusUseCase,
    ListReportsUseCase,
    GetDataSourceUseCase,
    ListDataSourcesUseCase
  ]
})
export class UseCasesModule {}
