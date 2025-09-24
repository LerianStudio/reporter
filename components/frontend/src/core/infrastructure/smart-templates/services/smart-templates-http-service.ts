import { Inject, Injectable, HttpService } from '@lerianstudio/sindarian-server'
import { LoggerAggregator, RequestIdRepository } from '@lerianstudio/lib-logs'
import { getServerSession } from 'next-auth'
import { nextAuthOptions } from '@/core/infrastructure/next-auth/next-auth-provider'
import { OtelTracerProvider } from '@/core/infrastructure/observability/otel-tracer-provider'
import { SpanStatusCode } from '@opentelemetry/api'
import { getIntl } from '@/lib/intl'
import { SmartTemplatesApiException } from '../exceptions/smart-templates-api-exceptions'
import { smartTemplatesApiMessages } from '../messages/messages'

@Injectable()
export class SmartTemplatesHttpService extends HttpService {
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

  private smartTemplatesCustomSpanName: string = 'smart-templates-api-request'

  protected async createDefaults() {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      'X-Request-Id': this.requestIdRepository.get()!
    }

    console.log('[INFO] - SmartTemplatesHttpService - Creating defaults', {})

    if (process.env.PLUGIN_AUTH_ENABLED === 'true') {
      console.log(
        '[INFO] - SmartTemplatesHttpService - PLUGIN_AUTH_ENABLED is',
        process.env.PLUGIN_AUTH_ENABLED
      )

      const session = await getServerSession(nextAuthOptions)
      if (session?.user?.access_token) {
        console.log(
          '[INFO] - SmartTemplatesHttpService - Session user access token',
          session.user.access_token
        )

        headers.Authorization = `Bearer ${session.user.access_token}`
      }
    }

    console.log('[INFO] - SmartTemplatesHttpService - Headers', headers)

    return {
      headers,
      baseUrl: process.env.PLUGIN_SMART_TEMPLATES_BASE_PATH
    }
  }

  protected onBeforeFetch(request: Request): void {
    this.logger.info('[INFO] - SmartTemplatesHttpService', {
      url: request.url,
      method: request.method,
      headers: {
        'X-Request-Id': request.headers.get('X-Request-Id'),
        'X-Organization-Id': request.headers.get('X-Organization-Id'),
        'Content-Type': request.headers.get('Content-Type')
      }
    })

    this.otelTracerProvider.startCustomSpan(this.smartTemplatesCustomSpanName)
  }

  protected onAfterFetch(request: Request, response: Response): void {
    this.logger.info('[INFO] - SmartTemplatesHttpService - Response', {
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
        'smart-templates.service': 'smart-templates',
        'organization.id': request.headers.get('X-Organization-Id') || 'unknown'
      },
      status: {
        code: response.ok ? SpanStatusCode.OK : SpanStatusCode.ERROR
      }
    })
  }

  protected async catch(request: Request, response: Response, error: any) {
    this.logger.error('[ERROR] - SmartTemplatesHttpService', {
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
        smartTemplatesApiMessages[
          error.code as keyof typeof smartTemplatesApiMessages
        ]

      if (!message) {
        this.logger.warn(
          '[ERROR] - SmartTemplatesHttpService - Error code not found',
          {
            url: request.url,
            method: request.method,
            status: response.status,
            response: error,
            errorCode: error.code
          }
        )

        throw new SmartTemplatesApiException(
          intl.formatMessage({
            id: 'error.smartTemplates.unknownError',
            defaultMessage: 'Unknown error occurred in Smart Templates service.'
          }),
          error.code,
          response.status
        )
      }

      throw new SmartTemplatesApiException(
        intl.formatMessage(message),
        error.code,
        response.status
      )
    }

    throw new SmartTemplatesApiException(
      intl.formatMessage({
        id: 'error.smartTemplates.unknownError',
        defaultMessage: 'Unknown error occurred in Smart Templates service.'
      }),
      'SMART_TEMPLATES_UNKNOWN_ERROR',
      response.status || 500
    )
  }

  /**
   * Handle file download streams with proper response handling
   * Override for binary file downloads from the Smart Templates API
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
}
