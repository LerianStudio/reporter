import { GenerateReportUseCase } from '@/core/application/use-cases/reports/generate-report-use-case'
import { ListReportsUseCase } from '@/core/application/use-cases/reports/list-reports-use-case'
import { GetReportStatusUseCase } from '@/core/application/use-cases/reports/get-report-status-use-case'
import { DownloadReportUseCase } from '@/core/application/use-cases/reports/download-report-use-case'
import { Container, ContainerModule } from '../../utils/di/container'

export const ReportUseCasesModule = new ContainerModule(
  (container: Container) => {
    // Report Use Cases registration
    container.bind<GenerateReportUseCase>(GenerateReportUseCase).toSelf()

    container.bind<ListReportsUseCase>(ListReportsUseCase).toSelf()

    container.bind<GetReportStatusUseCase>(GetReportStatusUseCase).toSelf()

    container.bind<DownloadReportUseCase>(DownloadReportUseCase).toSelf()
  }
)
