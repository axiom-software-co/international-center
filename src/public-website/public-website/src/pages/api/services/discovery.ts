// Service discovery endpoint for backend services
export const prerender = false;

import type { APIRoute } from 'astro';

export const GET: APIRoute = async ({ request }) => {
  try {
    const services = {
      'content-api': {
        name: 'content-api',
        url: process.env.API_BASE_URL || 'http://localhost:8080',
        health_endpoint: '/health',
        status: 'available',
        last_check: new Date().toISOString(),
      },
      'services-api': {
        name: 'services-api',
        url: process.env.SERVICES_API_URL || 'http://localhost:8081',
        health_endpoint: '/health',
        status: 'available',
        last_check: new Date().toISOString(),
      },
      'public-gateway': {
        name: 'public-gateway',
        url: process.env.PUBLIC_GATEWAY_URL || 'http://localhost:8082',
        health_endpoint: '/health',
        status: 'available',
        last_check: new Date().toISOString(),
      },
      'admin-gateway': {
        name: 'admin-gateway',
        url: process.env.ADMIN_GATEWAY_URL || 'http://localhost:8083',
        health_endpoint: '/health',
        status: 'available',
        last_check: new Date().toISOString(),
      },
    };

    // If Dapr is available, enhance with service mesh information
    const daprPort = process.env.DAPR_HTTP_PORT || '3500';
    let daprServices: any = {};
    
    try {
      const daprResponse = await fetch(`http://localhost:${daprPort}/v1.0/metadata`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      });
      
      if (daprResponse.ok) {
        const daprMetadata = await daprResponse.json();
        daprServices = {
          dapr_sidecar: {
            id: daprMetadata.id || 'website',
            version: daprMetadata.runtimeVersion || 'unknown',
            components: daprMetadata.components || [],
            actors: daprMetadata.actors || [],
          }
        };
      }
    } catch (daprError) {
      // Dapr not available, that's OK for now
    }

    const discoveryData = {
      timestamp: new Date().toISOString(),
      environment: process.env.NODE_ENV || 'development',
      service_mesh: Object.keys(daprServices).length > 0 ? 'dapr' : 'none',
      services: {
        ...services,
        ...daprServices,
      },
      total_services: Object.keys(services).length,
    };

    return new Response(JSON.stringify(discoveryData, null, 2), {
      status: 200,
      headers: {
        'Content-Type': 'application/json',
        'Cache-Control': 'no-cache, no-store, must-revalidate',
      },
    });

  } catch (error) {
    const errorData = {
      error: 'service_discovery_failed',
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