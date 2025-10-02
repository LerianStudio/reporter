import { Module, APP_FILTER } from '@lerianstudio/sindarian-server'
import { SmartTemplatesModule } from './smart-templates-module'
import { ControllerModule } from './controller-module'
import { UseCasesModule } from './use-cases-module'
import { LoggerModule } from './logger-module'
import { OtelModule } from './otel-module'
import { GlobalExceptionFilter } from '../smart-templates/exceptions/global-exception-filter'

@Module({
  imports: [
    LoggerModule,
    OtelModule,
    SmartTemplatesModule,
    UseCasesModule,
    ControllerModule
  ],
  providers: [
    {
      provide: APP_FILTER,
      useClass: GlobalExceptionFilter
    }
  ]
})
export class AppModule {}
