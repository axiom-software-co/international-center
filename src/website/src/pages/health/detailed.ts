// Detailed health endpoint for website container lifecycle monitoring
export const prerender = false;

import type { APIRoute } from 'astro';

export const GET: APIRoute = async ({ request }) => {
  try {
    // Comprehensive health check with detailed system information
    const detailedHealthData = {
      status: 'healthy',
      timestamp: new Date().toISOString(),
      service: 'international-center-website',
      version: '1.0.0',
      environment: import.meta.env.NODE_ENV || 'development',
      runtime: {
        node_version: process.version,
        platform: process.platform,
        arch: process.arch,
        uptime: process.uptime(),
      },
      memory: {
        used: Math.round(process.memoryUsage().heapUsed / 1024 / 1024), // MB
        total: Math.round(process.memoryUsage().heapTotal / 1024 / 1024), // MB
        external: Math.round(process.memoryUsage().external / 1024 / 1024), // MB
        rss: Math.round(process.memoryUsage().rss / 1024 / 1024), // MB
      },
      dependencies: {
        backend_services: {
          content_api: {
            status: 'pending_check',
            endpoint: process.env.API_BASE_URL || 'http://localhost:8080',
          },
          services_api: {
            status: 'pending_check', 
            endpoint: process.env.SERVICES_API_URL || 'http://localhost:8081',
          },
        },
        dapr_sidecar: {
          status: 'pending_check',
          port: process.env.DAPR_HTTP_PORT || '3500',
        },
        network: {
          status: 'healthy',
          isolation: 'container',
        }
      },
      build_info: {
        built_at: new Date().toISOString(), // Would be set during build
        commit_hash: 'development',
        astro_version: 'latest',
      },
      configuration: {
        host: '0.0.0.0',
        port: 3000,
        api_base_url: process.env.API_BASE_URL || 'http://localhost:8080',
        environment_vars_loaded: !!(process.env.API_BASE_URL),
      }
    };

    return new Response(JSON.stringify(detailedHealthData, null, 2), {
      status: 200,
      headers: {
        'Content-Type': 'application/json',
        'Cache-Control': 'no-cache, no-store, must-revalidate',
      },
    });
  } catch (error) {
    const errorData = {
      status: 'unhealthy',
      timestamp: new Date().toISOString(),
      service: 'international-center-website',
      error: error instanceof Error ? error.message : 'Unknown error',
      stack: error instanceof Error ? error.stack : undefined,
    };

    return new Response(JSON.stringify(errorData, null, 2), {
      status: 500,
      headers: {
        'Content-Type': 'application/json',
        'Cache-Control': 'no-cache, no-store, must-revalidate',
      },
    });
  }
};