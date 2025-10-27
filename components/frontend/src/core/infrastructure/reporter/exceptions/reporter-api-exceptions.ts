import { ApiException } from '@lerianstudio/sindarian-server'

/**
 * Custom exception for Reporter API errors
 * Handles errors from both templates and reports endpoints
 */
export class ReporterApiException extends ApiException {
  public readonly errorCode: string

  constructor(
    message: string,
    errorCode: string,
    status: number,
    response?: any
  ) {
    super(errorCode, 'Reporter API Error', message, status)
    this.errorCode = errorCode
    this.name = 'ReporterApiException'
  }
}
