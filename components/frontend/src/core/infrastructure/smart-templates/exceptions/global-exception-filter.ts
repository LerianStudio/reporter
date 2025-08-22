import { NextResponse } from 'next/server'
import {
  ArgumentsHost,
  ExceptionFilter,
  Inject,
  HttpStatus,
  Catch
} from '@lerianstudio/sindarian-server'
import { getIntl } from '@/lib/intl'
import { LoggerAggregator } from '@lerianstudio/lib-logs'
import { SmartTemplatesApiException } from './smart-templates-api-exceptions'

@Catch()
export class GlobalExceptionFilter implements ExceptionFilter {
  constructor(
    @Inject(LoggerAggregator) private readonly logger: LoggerAggregator
  ) {}

  async catch(exception: any, host: ArgumentsHost) {
    const intl = await getIntl()

    if (exception instanceof SmartTemplatesApiException) {
      this.logger.error(exception.message, exception.getResponse())
      return NextResponse.json(exception.getResponse(), {
        status: exception.getStatus()
      })
    }

    this.logger.error(`Unknown error`, exception)
    return NextResponse.json(
      {
        code: '0500',
        message: intl.formatMessage({
          id: 'error.midaz.unknowError',
          defaultMessage: 'Unknown error on Midaz.'
        })
      },
      {
        status: HttpStatus.INTERNAL_SERVER_ERROR
      }
    )
  }
}
