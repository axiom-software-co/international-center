// Build information endpoint for container lifecycle validation
export const prerender = false;

import type { APIRoute } from 'astro';

export const GET: APIRoute = async ({ request }) => {
  try {
    const buildInfo = {
      timestamp: new Date().toISOString(),
      service: 'international-center-website',
      build: {
        version: '1.0.0',
        built_at: new Date().toISOString(), // In real build, this would be set during CI/CD
        commit_hash: process.env.GIT_COMMIT || 'development',
        branch: process.env.GIT_BRANCH || 'main',
        build_number: process.env.BUILD_NUMBER || 'local',
      },
      runtime: {
        node_version: process.version,
        astro_version: 'latest', // Would be extracted from package.json
        framework: 'Astro + Vue',
        adapter: 'node-standalone',
        output_mode: 'hybrid',
      },
      container: {
        base_image: 'node:20-bookworm-slim',
        multi_stage: true,
        optimized: true,
        security: {
          non_root_user: true,
          minimal_dependencies: true,
        }
      },
      deployment: {
        environment: process.env.NODE_ENV || 'development',
        target: 'container',
        orchestration: 'pulumi',
        service_mesh: 'dapr',
      },
      features: {
        static_generation: true,
        server_side_rendering: true,
        api_routes: true,
        vue_components: true,
        tailwind_css: true,
        hot_reload: process.env.NODE_ENV === 'development',
      }
    };

    return new Response(JSON.stringify(buildInfo, null, 2), {
      status: 200,
      headers: {
        'Content-Type': 'application/json',
        'Cache-Control': 'public, max-age=3600', // Cache build info for 1 hour
      },
    });

  } catch (error) {
    const errorData = {
      error: 'build_info_failed',
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