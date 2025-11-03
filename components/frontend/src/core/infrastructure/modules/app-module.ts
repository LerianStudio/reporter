import { Module, APP_FILTER } from '@lerianstudio/sindarian-server'
import { ReporterModule } from './reporter-module'
import { ControllerModule } from './controller-module'
import { UseCasesModule } from './use-cases-module'
import { LoggerModule } from './logger-module'
import { OtelModule } from './otel-module'
import { GlobalExceptionFilter } from '../reporter/exceptions/global-exception-filter'

@Module({
  imports: [
    LoggerModule,
    OtelModule,
    ReporterModule,
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
