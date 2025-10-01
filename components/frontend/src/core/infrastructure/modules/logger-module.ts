import { Module, APP_INTERCEPTOR } from '@lerianstudio/sindarian-server'
import { ResolutionContext } from 'inversify'
import {
  LoggerAggregator,
  LoggerRepository,
  PinoLoggerRepository,
  RequestIdRepository
} from '@lerianstudio/lib-logs'
import { LoggerInterceptor } from '../logger/logger-interceptor'

@Module({
  providers: [
    {
      provide: APP_INTERCEPTOR,
      useClass: LoggerInterceptor
    },
    RequestIdRepository,
    {
      provide: LoggerRepository,
      useValue: new PinoLoggerRepository({
        debug: Boolean(process.env.ENABLE_DEBUG)
      })
    },
    {
      provide: LoggerAggregator,
      useFactory: (context: ResolutionContext) => {
        const loggerRepository = context.get<LoggerRepository>(LoggerRepository)
        return new LoggerAggregator(loggerRepository, {
          debug: Boolean(process.env.ENABLE_DEBUG)
        })
      }
    }
  ]
})
export class LoggerModule {}
