// Metrics endpoint for website observability integration
export const prerender = false;

import type { APIRoute } from 'astro';

export const GET: APIRoute = async ({ request }) => {
  try {
    // Get system metrics
    const memoryUsage = process.memoryUsage();
    const cpuUsage = process.cpuUsage();
    
    const metrics = {
      timestamp: new Date().toISOString(),
      service: 'international-center-website',
      environment: process.env.NODE_ENV || 'development',
      
      // System metrics
      system: {
        uptime: process.uptime(),
        memory: {
          heap_used: Math.round(memoryUsage.heapUsed / 1024 / 1024), // MB
          heap_total: Math.round(memoryUsage.heapTotal / 1024 / 1024), // MB
          rss: Math.round(memoryUsage.rss / 1024 / 1024), // MB
          external: Math.round(memoryUsage.external / 1024 / 1024), // MB
          heap_utilization: Math.round((memoryUsage.heapUsed / memoryUsage.heapTotal) * 100), // %
        },
        cpu: {
          user: cpuUsage.user,
          system: cpuUsage.system,
        },
        platform: {
          node_version: process.version,
          platform: process.platform,
          arch: process.arch,
        }
      },

      // Application metrics
      application: {
        requests: {
          total: 0, // Would be tracked in real implementation
          successful: 0,
          failed: 0,
          rate_per_minute: 0,
        },
        response_times: {
          avg: 0, // Would be tracked in real implementation
          p95: 0,
          p99: 0,
        },
        features: {
          api_proxy_calls: 0,
          health_checks: 0,
          static_files_served: 0,
        }
      },

      // Backend connectivity metrics
      backend_services: {
        content_api: {
          status: 'unknown',
          last_check: null,
          response_time: null,
          error_count: 0,
        },
        services_api: {
          status: 'unknown', 
          last_check: null,
          response_time: null,
          error_count: 0,
        }
      },

      // Container metrics
      container: {
        network: process.env.NETWORK_NAME || 'development-network',
        dapr_sidecar: !!(process.env.DAPR_HTTP_PORT),
        service_mesh_enabled: !!(process.env.DAPR_APP_ID),
      }
    };

    // Format as Prometheus-style metrics if requested
    const acceptHeader = request.headers.get('Accept');
    if (acceptHeader?.includes('text/plain')) {
      const prometheusMetrics = formatPrometheusMetrics(metrics);
      return new Response(prometheusMetrics, {
        status: 200,
        headers: {
          'Content-Type': 'text/plain; version=0.0.4',
          'Cache-Control': 'no-cache, no-store, must-revalidate',
        },
      });
    }

    return new Response(JSON.stringify(metrics, null, 2), {
      status: 200,
      headers: {
        'Content-Type': 'application/json',
        'Cache-Control': 'no-cache, no-store, must-revalidate',
      },
    });

  } catch (error) {
    const errorData = {
      error: 'metrics_collection_failed',
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

function formatPrometheusMetrics(metrics: any): string {
  const lines = [
    '# HELP website_uptime_seconds Total uptime in seconds',
    '# TYPE website_uptime_seconds counter',
    `website_uptime_seconds ${metrics.system.uptime}`,
    '',
    '# HELP website_memory_heap_used_bytes Memory heap used in bytes',
    '# TYPE website_memory_heap_used_bytes gauge',
    `website_memory_heap_used_bytes ${metrics.system.memory.heap_used * 1024 * 1024}`,
    '',
    '# HELP website_memory_heap_total_bytes Memory heap total in bytes',
    '# TYPE website_memory_heap_total_bytes gauge',
    `website_memory_heap_total_bytes ${metrics.system.memory.heap_total * 1024 * 1024}`,
    '',
    '# HELP website_memory_heap_utilization_percent Memory heap utilization percentage',
    '# TYPE website_memory_heap_utilization_percent gauge',
    `website_memory_heap_utilization_percent ${metrics.system.memory.heap_utilization}`,
    '',
  ];

  return lines.join('\n');
}