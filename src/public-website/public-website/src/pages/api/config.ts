// Configuration endpoint for website environment validation
export const prerender = false;

import type { APIRoute } from 'astro';

export const GET: APIRoute = async ({ request }) => {
  try {
    const configData = {
      timestamp: new Date().toISOString(),
      environment: process.env.NODE_ENV || 'development',
      service_name: 'international-center-website',
      version: '1.0.0',
      configuration: {
        server: {
          host: '0.0.0.0',
          port: 3000,
        },
        backend_apis: {
          content_api: {
            url: process.env.API_BASE_URL || 'http://localhost:8080',
            configured: !!(process.env.API_BASE_URL),
          },
          services_api: {
            url: process.env.SERVICES_API_URL || 'http://localhost:8081',
            configured: !!(process.env.SERVICES_API_URL),
          },
        },
        service_mesh: {
          dapr_enabled: !!(process.env.DAPR_HTTP_PORT),
          dapr_port: process.env.DAPR_HTTP_PORT || '3500',
          app_id: process.env.DAPR_APP_ID || 'website',
        },
        network: {
          network_name: process.env.NETWORK_NAME || 'development-network',
          container_runtime: 'docker', // or podman
        },
        build: {
          astro_mode: 'hybrid',
          node_adapter: 'standalone',
          build_target: 'container',
        },
        features: {
          api_proxy: true,
          health_checks: true,
          service_discovery: true,
          cors_enabled: true,
          development_features: process.env.NODE_ENV === 'development',
        }
      },
      environment_variables: {
        // Only expose non-sensitive env vars
        NODE_ENV: process.env.NODE_ENV,
        DAPR_HTTP_PORT: process.env.DAPR_HTTP_PORT,
        NETWORK_NAME: process.env.NETWORK_NAME,
        // Hide sensitive URLs in production
        ...(process.env.NODE_ENV === 'development' && {
          API_BASE_URL: process.env.API_BASE_URL,
          SERVICES_API_URL: process.env.SERVICES_API_URL,
        })
      }
    };

    return new Response(JSON.stringify(configData, null, 2), {
      status: 200,
      headers: {
        'Content-Type': 'application/json',
        'Cache-Control': 'no-cache, no-store, must-revalidate',
      },
    });

  } catch (error) {
    const errorData = {
      error: 'config_retrieval_failed',
      message: error instanceof Error ? error.message : 'Unknown error',
      timestamp: new Date().toISOString(),
    };

    return new Response(JSON.stringify(errorData), {
      status: 500,
      headers: {
        'Content-Type': 'application/json',
        'Cache-Control': 'no-cache, no-store, must-revalidate',
      },
    });
  }
};