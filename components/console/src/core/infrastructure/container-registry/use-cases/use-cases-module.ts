import { Container, ContainerModule } from '../../utils/di/container'
import { TemplateUseCasesModule } from './template-module'
import { ReportUseCasesModule } from './report-module'

export const UseCasesModule = new ContainerModule((container: Container) => {
  container.load(TemplateUseCasesModule)
  container.load(ReportUseCasesModule)
})
