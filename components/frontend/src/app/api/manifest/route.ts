import { NextResponse } from 'next/server'

/**
 * Plugin Manifest Endpoint for Smart Templates Plugin
 *
 * This endpoint is called by the console's plugin discovery system to register
 * and configure the Smart Templates plugin with NGINX proxy and authentication.
 *
 * The console will call:
 * - Development: http://localhost:8083/api/manifest
 * - Production: ${NGINX_BASE_PATH}/plugin-smart-templates-ui/api/manifest
 */

export async function GET() {
  try {
    const manifest = {
      name: 'plugin-reporter-ui',
      title: 'Reporter',
      description:
        'Reporter plugin for managing smart templates in the Midaz ecosystem.',
      version: '0.1.0',
      route:
        process.env.NEXT_PUBLIC_PLUGIN_UI_BASE_PATH || '/plugin-reporter-ui',
      entry: '/',
      healthcheck: '/api/admin/health',
      host:
        process.env.NODE_ENV === 'development'
          ? 'http://localhost:8083'
          : process.env.PLUGIN_SMART_TEMPLATES_UI_BASE_URL ||
            'http://plugin-smart-templates-ui:8083',
      icon: 'LayoutTemplate',
      enabled: true,
      author: 'Lerian Studio'
    }

    console.log('[INFO] - Plugin Smart Templates UI Manifest requested', {
      manifest,
      nodeEnv: process.env.NODE_ENV,
      timestamp: new Date().toISOString()
    })

    return NextResponse.json(manifest)
  } catch (error: any) {
    console.error(
      '[ERROR] - Failed to generate plugin Smart Templates UI manifest:',
      error
    )

    return NextResponse.json(
      {
        error: 'Failed to generate plugin Smart Templates UI manifest',
        details: error?.message
      },
      { status: 500 }
    )
  }
}
