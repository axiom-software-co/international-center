// API Explorer endpoint for development features
export const prerender = false;

import type { APIRoute } from 'astro';

export const GET: APIRoute = async ({ request }) => {
  // Only available in development
  if (process.env.NODE_ENV !== 'development') {
    return new Response(JSON.stringify({
      error: 'not_available',
      message: 'Development features only available in development mode',
    }), {
      status: 404,
      headers: {
        'Content-Type': 'application/json',
      },
    });
  }

  try {
    const apiExplorer = {
      timestamp: new Date().toISOString(),
      feature: 'api-explorer',
      status: 'available',
      environment: 'development',
      available_apis: {
        website_health: {
          endpoint: '/health',
          method: 'GET',
          description: 'Website health check',
        },
        detailed_health: {
          endpoint: '/health/detailed',
          method: 'GET', 
          description: 'Detailed website health information',
        },
        content_api_proxy: {
          endpoint: '/api/content/health',
          method: 'GET',
          description: 'Proxy to Content API health endpoint',
        },
        services_api_proxy: {
          endpoint: '/api/services/health',
          method: 'GET',
          description: 'Proxy to Services API health endpoint',
        },
        service_discovery: {
          endpoint: '/api/services/discovery',
          method: 'GET',
          description: 'Service discovery information',
        },
        network_status: {
          endpoint: '/api/network/status',
          method: 'GET',
          description: 'Network connectivity status',
        },
        configuration: {
          endpoint: '/api/config',
          method: 'GET',
          description: 'Website configuration information',
        },
        build_info: {
          endpoint: '/build-info',
          method: 'GET',
          description: 'Build and deployment information',
        }
      },
      development_features: {
        hot_reload: {
          endpoint: '/dev/hot-reload',
          method: 'GET',
          description: 'Hot reload status',
        },
        component_preview: {
          endpoint: '/dev/component-preview',
          method: 'GET',
          description: 'Component preview system',
        }
      }
    };

    return new Response(JSON.stringify(apiExplorer, null, 2), {
      status: 200,
      headers: {
        'Content-Type': 'application/json',
        'Cache-Control': 'no-cache, no-store, must-revalidate',
      },
    });

  } catch (error) {
    const errorData = {
      error: 'api_explorer_failed',
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