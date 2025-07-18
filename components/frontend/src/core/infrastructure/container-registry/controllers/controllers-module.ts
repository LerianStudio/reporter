import { Container, ContainerModule } from '../../utils/di/container'
import { TemplateController } from '@/core/application/controllers/template-controller'
import { ReportController } from '@/core/application/controllers/report-controller'

export const ControllersModule = new ContainerModule((container: Container) => {
  // Controllers registration
  container.bind<TemplateController>(TemplateController).toSelf()

  container.bind<ReportController>(ReportController).toSelf()
})
