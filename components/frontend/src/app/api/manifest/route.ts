import { NextResponse } from 'next/server'

/**
 * Manifest Endpoint for Reporter
 *
 * This endpoint is called by the console's discovery system to register
 * and configure the Reporter application with NGINX proxy and authentication.
 *
 * The console will call:
 * - Development: http://localhost:8083/api/manifest
 * - Production: ${NGINX_BASE_PATH}/reporter-ui/api/manifest
 */

export async function GET() {
  try {
    const manifest = {
      name: 'reporter-ui',
      title: 'Reporter',
      description:
        'Reporter application for managing templates in the Midaz ecosystem.',
      version: '0.1.0',
      route: process.env.NEXT_PUBLIC_REPORTER_UI_BASE_PATH || '/reporter-ui',
      entry: '/',
      healthcheck: '/api/admin/health',
      host:
        process.env.NODE_ENV === 'development'
          ? 'http://localhost:8083'
          : process.env.REPORTER_UI_BASE_URL || 'http://reporter-ui:8083',
      icon: 'LayoutTemplate',
      enabled: true,
      author: 'Lerian Studio'
    }

    console.log('[INFO] - Reporter UI Manifest requested', {
      manifest,
      nodeEnv: process.env.NODE_ENV,
      timestamp: new Date().toISOString()
    })

    return NextResponse.json(manifest)
  } catch (error: any) {
    console.error('[ERROR] - Failed to generate Reporter UI manifest:', error)

    return NextResponse.json(
      {
        error: 'Failed to generate Reporter UI manifest',
        details: error?.message
      },
      { status: 500 }
    )
  }
}
