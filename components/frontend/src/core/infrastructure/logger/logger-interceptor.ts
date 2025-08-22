import { LoggerAggregator, RequestIdRepository } from '@lerianstudio/lib-logs'
import {
  CallHandler,
  ExecutionContext,
  Inject,
  Interceptor
} from '@lerianstudio/sindarian-server'
import { NextRequest } from 'next/server'

export class LoggerInterceptor implements Interceptor {
  constructor(
    @Inject(RequestIdRepository)
    private requestIdRepository: RequestIdRepository,
    @Inject(LoggerAggregator) private logger: LoggerAggregator
  ) {}

  async intercept(context: ExecutionContext, next: CallHandler): Promise<any> {
    const request = context.switchToHttp().getRequest<NextRequest>()
    const traceId = this.requestIdRepository.generate()
    this.requestIdRepository.set(traceId)

    // const body =
    //   request.method !== 'GET' && request.method !== 'DELETE'
    //     ? { body: request.body }
    //     : {}

    return await this.logger.runWithContext(
      request.url,
      request.method,
      {
        requestId: this.requestIdRepository.get(),
        handler: `${context.getClass().name}.${context.getHandler().name}`
        // ...body
      },
      () => next.handle()
    )
  }
}
