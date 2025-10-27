/** @type {import('next').NextConfig} */

const basePath = process.env.NEXT_PUBLIC_REPORTER_UI_BASE_PATH

const nextConfig = {
  basePath: process.env.NEXT_PUBLIC_REPORTER_UI_BASE_PATH,
  assetPrefix: process.env.NEXT_PUBLIC_REPORTER_UI_BASE_PATH,
  trailingSlash: true,
  skipTrailingSlashRedirect: true,
  reactStrictMode: false,
  logging: {
    fetches: {
      fullUrl: true
    }
  },
  env: {
    BASE_PATH: process.env.BASE_PATH
  },
  headers: async () => {
    return [
      {
        source: '/api/:path*',
        headers: [
          { key: 'Access-Control-Allow-Credentials', value: 'true' },
          { key: 'Access-Control-Allow-Origin', value: '*' }, // replace this your actual origin
          {
            key: 'Access-Control-Allow-Methods',
            value: 'GET,DELETE,PATCH,POST,PUT'
          },
          {
            key: 'Access-Control-Allow-Headers',
            value:
              'X-CSRF-Token, X-Requested-With, Accept, Accept-Version, Content-Length, Content-MD5, Content-Type, Date, X-Api-Version'
          }
        ]
      }
    ]
  },
  redirects: async () => {
    return basePath
      ? [
          {
            source: '/',
            destination: basePath,
            basePath: false,
            permanent: false
          }
        ]
      : []
  },
  images: {
    dangerouslyAllowSVG: true,
    contentSecurityPolicy: "default-src 'self'; script-src 'none'; sandbox;",
    contentDispositionType: 'attachment',

    remotePatterns: [
      {
        protocol: 'http',
        hostname: '**',
        pathname: '**'
      },
      {
        protocol: 'https',
        hostname: '**',
        pathname: '**'
      }
    ]
  },
  compiler: {
    reactRemoveProperties:
      process.env.NODE_ENV === 'production'
        ? { properties: ['^data-testid$'] }
        : false
  },
  webpack: (config, { isServer }) => {
    config.resolve.fallback = {
      ...config.resolve.fallback,
      worker_threads: false,
      pino: false
    }

    return config
  },

  transpilePackages: [
    '@lerianstudio/console-layout',
    '@lerianstudio/sindarian-ui'
  ],

  serverExternalPackages: [
    'pino',
    'pino-pretty',
    '@opentelemetry/instrumentation',
    '@opentelemetry/api',
    '@opentelemetry/api-logs',
    '@opentelemetry/exporter-logs-otlp-http',
    '@opentelemetry/exporter-metrics-otlp-http',
    '@opentelemetry/exporter-trace-otlp-http',
    '@opentelemetry/instrumentation',
    '@opentelemetry/instrumentation-http',
    '@opentelemetry/instrumentation-pino',
    '@opentelemetry/instrumentation-runtime-node',
    '@opentelemetry/resources',
    '@opentelemetry/sdk-logs',
    '@opentelemetry/sdk-metrics',
    '@opentelemetry/sdk-node',
    '@opentelemetry/sdk-trace-base',
    '@opentelemetry/instrumentation-undici'
  ]
}

export default nextConfig
