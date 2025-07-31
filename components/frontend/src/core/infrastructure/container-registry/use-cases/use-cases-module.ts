import { Container, ContainerModule } from '../../utils/di/container'
import { TemplateUseCasesModule } from './template-module'
import { ReportUseCasesModule } from './report-module'
import { DataSourceUseCasesModule } from './data-source-module'

export const UseCasesModule = new ContainerModule((container: Container) => {
  container.load(TemplateUseCasesModule)
  container.load(ReportUseCasesModule)
  container.load(DataSourceUseCasesModule)
})
