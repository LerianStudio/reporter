import { CreateTemplateUseCase } from '@/core/application/use-cases/templates/create-template-use-case'
import { DeleteTemplateUseCase } from '@/core/application/use-cases/templates/delete-template-use-case'
import { GetTemplateUseCase } from '@/core/application/use-cases/templates/get-template-use-case'
import { ListTemplatesUseCase } from '@/core/application/use-cases/templates/list-templates-use-case'
import { UpdateTemplateUseCase } from '@/core/application/use-cases/templates/update-template-use-case'
import { Container, ContainerModule } from '../../utils/di/container'

export const TemplateUseCasesModule = new ContainerModule(
  (container: Container) => {
    // Use Cases registration
    container
      .bind<CreateTemplateUseCase>(CreateTemplateUseCase)
      .toSelf()
      .inTransientScope()

    container
      .bind<ListTemplatesUseCase>(ListTemplatesUseCase)
      .toSelf()
      .inTransientScope()

    container
      .bind<GetTemplateUseCase>(GetTemplateUseCase)
      .toSelf()
      .inTransientScope()

    container
      .bind<UpdateTemplateUseCase>(UpdateTemplateUseCase)
      .toSelf()
      .inTransientScope()

    container
      .bind<DeleteTemplateUseCase>(DeleteTemplateUseCase)
      .toSelf()
      .inTransientScope()
  }
)
