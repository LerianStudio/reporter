import { LoggerAggregator } from '@lerianstudio/lib-logs'
import { container } from '@/core/infrastructure/container-registry/container-registry'
import { HttpStatus } from '@/lib/http'
import { getIntl } from '@/lib/intl'
import { SmartTemplatesApiException } from '@/core/infrastructure/smart-templates/exceptions/smart-templates-api-exceptions'

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
  console.log('apiErrorHandler', error)
  if (error instanceof SmartTemplatesApiException) {
    logger.error(`SmartTemplatesApiException`, errorMetadata)
    return {
      message: error.message,
      status: error.getStatus()
    }
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
