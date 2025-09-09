// Services API health proxy endpoint
export const prerender = false;

import type { APIRoute } from 'astro';

export const GET: APIRoute = async ({ request }) => {
  try {
    const servicesApiUrl = process.env.SERVICES_API_URL || 'http://localhost:8081';
    const healthEndpoint = `${servicesApiUrl}/health`;

    // Proxy request to Services API via Dapr service invocation if available
    const daprPort = process.env.DAPR_HTTP_PORT || '3500';
    const daprEndpoint = `http://localhost:${daprPort}/v1.0/invoke/services-api/method/health`;
    
    let response: Response;
    let proxyMethod = 'direct';
    
    try {
      // Try Dapr service invocation first
      response = await fetch(daprEndpoint, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      });
      proxyMethod = 'dapr';
    } catch (daprError) {
      // Fallback to direct API call
      response = await fetch(healthEndpoint, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      });
      proxyMethod = 'direct';
    }

    if (!response.ok) {
      throw new Error(`Services API health check failed with status: ${response.status}`);
    }

    const healthData = await response.json();
    
    // Enhance with proxy information
    const enhancedData = {
      ...healthData,
      proxy: {
        method: proxyMethod,
        timestamp: new Date().toISOString(),
        endpoint_used: proxyMethod === 'dapr' ? daprEndpoint : healthEndpoint,
      }
    };

    return new Response(JSON.stringify(enhancedData), {
      status: response.status,
      headers: {
        'Content-Type': 'application/json',
        'Cache-Control': 'no-cache, no-store, must-revalidate',
      },
    });

  } catch (error) {
    const errorData = {
      status: 'error',
      service: 'services-api-proxy',
      message: error instanceof Error ? error.message : 'Unknown error',
      timestamp: new Date().toISOString(),
      proxy: {
        method: 'failed',
        attempted_endpoints: [
          `http://localhost:${process.env.DAPR_HTTP_PORT || '3500'}/v1.0/invoke/services-api/method/health`,
          `${process.env.SERVICES_API_URL || 'http://localhost:8081'}/health`,
        ],
      }
    };

    return new Response(JSON.stringify(errorData), {
      status: 503,
      headers: {
        'Content-Type': 'application/json',
        'Cache-Control': 'no-cache, no-store, must-revalidate',
      },
    });
  }
};