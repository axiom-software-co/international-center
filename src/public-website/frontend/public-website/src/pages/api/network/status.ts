// Network connectivity validation endpoint
export const prerender = false;

import type { APIRoute } from 'astro';

export const GET: APIRoute = async ({ request }) => {
  try {
    const networkStatus = {
      timestamp: new Date().toISOString(),
      container_network: {
        isolation: 'enabled',
        network_name: process.env.NETWORK_NAME || 'development-network',
        status: 'connected',
      },
      connectivity: {
        backend_services: {
          content_api: await checkServiceConnectivity('content-api', process.env.API_BASE_URL || 'http://localhost:8080'),
          services_api: await checkServiceConnectivity('services-api', process.env.SERVICES_API_URL || 'http://localhost:8081'),
        },
        dapr_sidecar: await checkDaprConnectivity(),
        external_network: await checkExternalConnectivity(),
      },
      network_configuration: {
        host: '0.0.0.0',
        port: 3000,
        environment: process.env.NODE_ENV || 'development',
      }
    };

    return new Response(JSON.stringify(networkStatus, null, 2), {
      status: 200,
      headers: {
        'Content-Type': 'application/json',
        'Cache-Control': 'no-cache, no-store, must-revalidate',
      },
    });

  } catch (error) {
    const errorData = {
      error: 'network_status_check_failed',
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

async function checkServiceConnectivity(serviceName: string, serviceUrl: string) {
  try {
    // Use contract-generated health client for type-safe health checks
    const { apiClient } = await import('../../../lib/api-client');
    
    const healthResponse = await apiClient.getHealth();
    
    return {
      service: serviceName,
      url: serviceUrl,
      status: healthResponse.status === 'healthy' ? 'connected' : 'unreachable',
      response_code: 200,
      health_details: healthResponse,
      client_type: 'contract-generated',
      last_checked: new Date().toISOString(),
    };
  } catch (error) {
    return {
      service: serviceName,
      url: serviceUrl,
      status: 'unreachable',
      error: error instanceof Error ? error.message : 'Contract client error',
      client_type: 'contract-generated',
      last_checked: new Date().toISOString(),
    };
  }
}

async function checkDaprConnectivity() {
  try {
    const daprPort = process.env.DAPR_HTTP_PORT || '3500';
    const response = await fetch(`http://localhost:${daprPort}/v1.0/healthz`, {
      method: 'GET',
      signal: AbortSignal.timeout(3000), // 3 second timeout
    });
    
    return {
      status: response.ok ? 'connected' : 'unreachable',
      port: daprPort,
      response_code: response.status,
      last_checked: new Date().toISOString(),
    };
  } catch (error) {
    return {
      status: 'unreachable',
      port: process.env.DAPR_HTTP_PORT || '3500',
      error: error instanceof Error ? error.message : 'Unknown error',
      last_checked: new Date().toISOString(),
    };
  }
}

async function checkExternalConnectivity() {
  try {
    // Simple external connectivity check (to a reliable service)
    const response = await fetch('https://httpbin.org/status/200', {
      method: 'GET',
      signal: AbortSignal.timeout(5000), // 5 second timeout
    });
    
    return {
      status: response.ok ? 'connected' : 'limited',
      last_checked: new Date().toISOString(),
    };
  } catch (error) {
    return {
      status: 'offline',
      error: error instanceof Error ? error.message : 'Unknown error',
      last_checked: new Date().toISOString(),
    };
  }
}