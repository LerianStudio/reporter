import { Container, ContainerModule } from '../../utils/di/container'
import { MidazHttpService } from '../../midaz/services/midaz-http-service'

export const MidazModule = new ContainerModule((container: Container) => {
  container.bind<MidazHttpService>(MidazHttpService).toSelf()
})
