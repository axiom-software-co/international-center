// Component preview system for development
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
    const componentPreview = {
      timestamp: new Date().toISOString(),
      feature: 'component-preview',
      status: 'active',
      environment: 'development',
      available_components: {
        ui_components: {
          navigation: ['NavigationMenu', 'Footer', 'MobileNav'],
          forms: ['VolunteerForm', 'LargeDonationForm', 'ContactForm'],
          content: ['NewsContent', 'EventContent', 'ServicePage'],
          blocks: ['Testimonials', 'ComprehensiveCare', 'FacilityDevelopment'],
        },
        page_components: {
          layouts: ['Layout'],
          pages: ['HomePage', 'AboutPage', 'ServicesPage'],
        }
      },
      preview_system: {
        hot_reload: true,
        style_injection: true,
        props_editor: false, // Future enhancement
        responsive_preview: false, // Future enhancement
      },
      development_tools: {
        vue_devtools: true,
        astro_devtools: true,
        tailwind_debug: true,
      }
    };

    return new Response(JSON.stringify(componentPreview, null, 2), {
      status: 200,
      headers: {
        'Content-Type': 'application/json',
        'Cache-Control': 'no-cache, no-store, must-revalidate',
      },
    });

  } catch (error) {
    const errorData = {
      error: 'component_preview_failed',
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