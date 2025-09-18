import { Module } from '@lerianstudio/sindarian-server'
import { OtelTracerProvider } from '../observability/otel-tracer-provider'

@Module({
  providers: [OtelTracerProvider]
})
export class OtelModule {}
