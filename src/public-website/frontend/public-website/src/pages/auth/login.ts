// Authentication login endpoint
export const prerender = false;

import type { APIRoute } from 'astro';

export const GET: APIRoute = async ({ request }) => {
  try {
    // In development, return auth flow information
    const authInfo = {
      timestamp: new Date().toISOString(),
      auth_system: 'development',
      environment: process.env.NODE_ENV || 'development',
      
      flow: {
        type: 'oauth2_placeholder',
        status: 'not_implemented',
        providers: ['auth0', 'google', 'azure_ad'],
        redirect_uri: `${new URL(request.url).origin}/auth/callback`,
      },
      
      development: {
        bypass: true,
        test_users: ['test@example.com', 'admin@example.com'],
        session_management: 'disabled',
      },
      
      backend_integration: {
        user_service: process.env.AUTH_SERVICE_URL || 'http://localhost:8084',
        jwt_validation: 'disabled',
        role_management: 'placeholder',
      }
    };

    // Set CORS headers for auth endpoints
    return new Response(JSON.stringify(authInfo, null, 2), {
      status: 200,
      headers: {
        'Content-Type': 'application/json',
        'Cache-Control': 'no-cache, no-store, must-revalidate',
        'Access-Control-Allow-Origin': '*',
        'Access-Control-Allow-Methods': 'GET, POST, OPTIONS',
        'Access-Control-Allow-Headers': 'Content-Type, Authorization',
      },
    });

  } catch (error) {
    const errorData = {
      error: 'auth_login_failed',
      message: error instanceof Error ? error.message : 'Unknown error',
      timestamp: new Date().toISOString(),
    };

    return new Response(JSON.stringify(errorData), {
      status: 500,
      headers: {
        'Content-Type': 'application/json',
        'Cache-Control': 'no-cache, no-store, must-revalidate',
        'Access-Control-Allow-Origin': '*',
      },
    });
  }
};

// Handle CORS preflight requests
export const OPTIONS: APIRoute = async ({ request }) => {
  return new Response(null, {
    status: 200,
    headers: {
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Methods': 'GET, POST, OPTIONS',
      'Access-Control-Allow-Headers': 'Content-Type, Authorization',
      'Access-Control-Max-Age': '86400',
    },
  });
};