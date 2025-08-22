import { ApiException } from '@lerianstudio/sindarian-server'

/**
 * Custom exception for Smart Templates API errors
 * Handles errors from both templates and reports endpoints
 */
export class SmartTemplatesApiException extends ApiException {
  public readonly errorCode: string

  constructor(
    message: string,
    errorCode: string,
    status: number,
    response?: any
  ) {
    super(errorCode, 'Smart Templates API Error', message, status)
    this.errorCode = errorCode
    this.name = 'SmartTemplatesApiException'
  }
}
