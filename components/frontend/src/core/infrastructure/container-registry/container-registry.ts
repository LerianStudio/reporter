import 'reflect-metadata'

import { Container } from '../utils/di/container'

import { LoggerModule } from './logger/logger-module'
import { UseCasesModule } from './use-cases/use-cases-module'
import { ControllersModule } from './controllers/controllers-module'
import { MidazModule } from './midaz/midaz-module'
import { SmartTemplatesModule } from './smart-templates/smart-templates-module'
import { OtelModule } from './observability/otel-module'

export const container = new Container()

container.load(ControllersModule)
container.load(LoggerModule)
container.load(MidazModule)
container.load(OtelModule)
container.load(SmartTemplatesModule)
container.load(UseCasesModule)
