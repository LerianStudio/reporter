import { LoggerAggregator } from '@lerianstudio/lib-logs'
import { container } from '@/core/infrastructure/container-registry/container-registry'
import { HttpStatus } from '@/lib/http'
import { getIntl } from '@/lib/intl'

export interface ErrorResponse {
  message: string
  status: number
}

export async function apiErrorHandler(error: any): Promise<ErrorResponse> {
  const intl = await getIntl()
  const logger = container.get<LoggerAggregator>(LoggerAggregator)

  const errorMetadata = {
    errorType: error.constructor.name,
    originalMessage: error.message
  }

  logger.error(`Unknown error`, errorMetadata)
  return {
    message: intl.formatMessage({
      id: 'error.midaz.unknowError',
      defaultMessage: 'Unknown error on Midaz.'
    }),
    status: HttpStatus.INTERNAL_SERVER_ERROR
  }
}
