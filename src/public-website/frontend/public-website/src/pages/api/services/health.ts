// Contract-compliant services API health endpoint using generated client
export const prerender = false;

import type { APIRoute } from 'astro';

export const GET: APIRoute = async ({ request }) => {
  try {
    // Use contract-generated API client for type-safe health checks
    const { apiClient } = await import('../../../lib/api-client');
    
    const healthResponse = await apiClient.getHealth();
    
    // Enhance with contract client information
    const enhancedData = {
      ...healthResponse,
      proxy: {
        method: 'contract-client',
        timestamp: new Date().toISOString(),
        client_version: '1.0.0',
        contract_compliant: true,
        service: 'services-api',
      }
    };

    return new Response(JSON.stringify(enhancedData), {
      status: 200,
      headers: {
        'Content-Type': 'application/json',
        'Cache-Control': 'no-cache, no-store, must-revalidate',
      },
    });

  } catch (error) {
    // Contract-compliant error response
    const errorData = {
      status: 'error',
      service: 'services-api-contract-proxy',
      message: error instanceof Error ? error.message : 'Unknown error',
      timestamp: new Date().toISOString(),
      proxy: {
        method: 'contract-client-failed',
        client_version: '1.0.0',
        contract_compliant: true,
      },
      error: {
        code: 'HEALTH_CHECK_FAILED',
        details: error instanceof Error ? error.stack : 'Unknown error details',
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