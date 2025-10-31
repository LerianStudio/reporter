import { Inject, Injectable, HttpService } from '@lerianstudio/sindarian-server'
import { LoggerAggregator, RequestIdRepository } from '@lerianstudio/lib-logs'
import { getServerSession } from 'next-auth'
import { nextAuthOptions } from '@/core/infrastructure/next-auth/next-auth-provider'
import { OtelTracerProvider } from '@/core/infrastructure/observability/otel-tracer-provider'
import { SpanStatusCode } from '@opentelemetry/api'
import { getIntl } from '@/lib/intl'
import { ReporterApiException } from '../exceptions/reporter-api-exceptions'
import { reporterApiMessages } from '../messages/messages'

@Injectable()
export class ReporterHttpService extends HttpService {
  constructor(
    @Inject(LoggerAggregator)
    private readonly logger: LoggerAggregator,
    @Inject(RequestIdRepository)
    private readonly requestIdRepository: RequestIdRepository,
    @Inject(OtelTracerProvider)
    private readonly otelTracerProvider: OtelTracerProvider
  ) {
    super()
  }

  private reporterCustomSpanName: string = 'reporter-api-request'

  protected async createDefaults() {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      'X-Request-Id': this.requestIdRepository.get()!
    }

    if (process.env.PLUGIN_AUTH_ENABLED === 'true') {
      const session = await getServerSession(nextAuthOptions)
      if (session?.user?.access_token) {
        headers.Authorization = `Bearer ${session.user.access_token}`
      }
    }

    return {
      headers,
      baseUrl: process.env.REPORTER_BASE_PATH
    }
  }

  protected onBeforeFetch(request: Request): void {
    this.logger.info('[INFO] - ReporterHttpService', {
      url: request.url,
      method: request.method,
      headers: {
        'X-Request-Id': request.headers.get('X-Request-Id'),
        'X-Organization-Id': request.headers.get('X-Organization-Id'),
        'Content-Type': request.headers.get('Content-Type')
      }
    })

    this.otelTracerProvider.startCustomSpan(this.reporterCustomSpanName)
  }

  protected onAfterFetch(request: Request, response: Response): void {
    this.logger.info('[INFO] - ReporterHttpService - Response', {
      url: request.url,
      method: request.method,
      status: response.status,
      statusText: response.statusText
    })

    this.otelTracerProvider.endCustomSpan({
      attributes: {
        'http.url': request.url,
        'http.method': request.method,
        'http.status_code': response.status,
        'reporter.service': 'reporter',
        'organization.id': request.headers.get('X-Organization-Id') || 'unknown'
      },
      status: {
        code: response.ok ? SpanStatusCode.OK : SpanStatusCode.ERROR
      }
    })
  }

  protected async catch(request: Request, response: Response, error: any) {
    this.logger.error('[ERROR] - ReporterHttpService', {
      url: request.url,
      method: request.method,
      status: response.status,
      response: error,
      headers: {
        'X-Request-Id': request.headers.get('X-Request-Id'),
        'X-Organization-Id': request.headers.get('X-Organization-Id')
      }
    })

    const intl = await getIntl()

    if (error?.code) {
      const message =
        reporterApiMessages[error.code as keyof typeof reporterApiMessages]

      if (!message) {
        this.logger.warn(
          '[ERROR] - ReporterHttpService - Error code not found',
          {
            url: request.url,
            method: request.method,
            status: response.status,
            response: error,
            errorCode: error.code
          }
        )

        throw new ReporterApiException(
          intl.formatMessage({
            id: 'error.reporter.unknownError',
            defaultMessage: 'Unknown error occurred in Reporter service.'
          }),
          error.code,
          response.status
        )
      }

      throw new ReporterApiException(
        intl.formatMessage(message),
        error.code,
        response.status
      )
    }

    throw new ReporterApiException(
      intl.formatMessage({
        id: 'error.reporter.unknownError',
        defaultMessage: 'Unknown error occurred in Reporter service.'
      }),
      'REPORTER_UNKNOWN_ERROR',
      response.status || 500
    )
  }

  /**
   * Handle file download streams with proper response handling
   * Override for binary file downloads from the Reporter API
   */
  async getFileStream(
    url: URL | string,
    options: RequestInit = {}
  ): Promise<Response> {
    const defaults = await this.createDefaults()

    const headers = { ...defaults.headers }
    delete headers['Content-Type']

    const request = new Request(new URL(url, defaults.baseUrl), {
      ...defaults,
      ...options,
      method: 'GET',
      headers: {
        ...headers,
        ...options.headers
      }
    })

    this.onBeforeFetch(request)

    const response = await fetch(request)

    this.onAfterFetch(request, response)

    if (!response.ok) {
      const error = await response
        .json()
        .catch(() => ({ message: 'Download failed' }))
      await this.catch(request, response, error)
    }

    return response
  }

  /**
   * Handle text responses for file content downloads
   * Returns the response content as text instead of parsing as JSON
   */
  async getText(url: URL | string, options: RequestInit = {}): Promise<string> {
    const defaults = await this.createDefaults()

    const request = new Request(new URL(url, defaults.baseUrl), {
      ...defaults,
      ...options,
      method: 'GET',
      headers: {
        ...defaults.headers,
        ...options.headers
      }
    })

    this.onBeforeFetch(request)

    const response = await fetch(request)

    this.onAfterFetch(request, response)

    if (!response.ok) {
      const error = await response
        .json()
        .catch(() => ({ message: 'Download failed' }))
      await this.catch(request, response, error)
    }

    return await response.text()
  }

  /**
   * Handle binary responses for file content downloads
   * Returns the response content as ArrayBuffer for binary data
   */
  async getBinary(
    url: URL | string,
    options: RequestInit = {}
  ): Promise<ArrayBuffer> {
    const defaults = await this.createDefaults()

    const request = new Request(new URL(url, defaults.baseUrl), {
      ...defaults,
      ...options,
      method: 'GET',
      headers: {
        ...defaults.headers,
        ...options.headers
      }
    })

    this.onBeforeFetch(request)

    const response = await fetch(request)

    this.onAfterFetch(request, response)

    if (!response.ok) {
      const error = await response
        .json()
        .catch(() => ({ message: 'Download failed' }))
      await this.catch(request, response, error)
    }

    return await response.arrayBuffer()
  }
}
