import { ListDataSourcesUseCase } from '@/core/application/use-cases/data-sources/list-data-sources-use-case'
import { GetDataSourceDetailsUseCase } from '@/core/application/use-cases/data-sources/get-data-source-details-use-case'
import { Container, ContainerModule } from '../../utils/di/container'

export const DataSourceUseCasesModule = new ContainerModule(
  (container: Container) => {
    // Use Cases registration
    container
      .bind<ListDataSourcesUseCase>(ListDataSourcesUseCase)
      .toSelf()
      .inTransientScope()

    container
      .bind<GetDataSourceDetailsUseCase>(GetDataSourceDetailsUseCase)
      .toSelf()
      .inTransientScope()
  }
) 