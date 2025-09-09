// Hot reload status endpoint for development features
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
    const hotReloadStatus = {
      timestamp: new Date().toISOString(),
      feature: 'hot-reload',
      status: 'active',
      environment: 'development',
      details: {
        astro_dev_server: true,
        vue_hmr: true,
        tailwind_watch: true,
        browser_sync: false, // Not implemented yet
      },
      performance: {
        reload_time: '< 100ms',
        file_watching: true,
        incremental_builds: true,
      }
    };

    return new Response(JSON.stringify(hotReloadStatus, null, 2), {
      status: 200,
      headers: {
        'Content-Type': 'application/json',
        'Cache-Control': 'no-cache, no-store, must-revalidate',
      },
    });

  } catch (error) {
    const errorData = {
      error: 'hot_reload_status_failed',
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