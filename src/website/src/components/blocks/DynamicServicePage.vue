<template>
  <div>
    <!-- Error State -->
    <div v-if="error" class="text-center py-12">
      <div class="max-w-md mx-auto">
        <h3 class="text-lg font-semibold text-gray-900 mb-2">Service Temporarily Unavailable</h3>
        <p class="text-gray-600 mb-4">
          We're experiencing technical difficulties. Please try again later.
        </p>
        <a
          href="/services"
          class="inline-block px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
        >
          Browse All Services
        </a>
      </div>
    </div>

    <!-- Loading State -->
    <div v-else-if="isLoading">
      <section class="pb-8">
        <!-- Breadcrumb Loading -->
        <div class="bg-gray-50 py-6">
          <div class="container">
            <div class="animate-pulse">
              <div class="h-4 bg-gray-300 rounded w-64 mb-4"></div>
              <div class="h-8 bg-gray-300 rounded w-96 mb-2"></div>
              <div class="h-4 bg-gray-300 rounded w-80"></div>
            </div>
          </div>
        </div>

        <div class="container service-page-container">
          <div class="mt-8 grid gap-12 md:grid-cols-12 md:gap-8">
            <!-- Main Content Loading -->
            <div class="order-2 md:order-none md:col-span-7 md:col-start-1 lg:col-span-8">
              <article class="prose dark:prose-invert mx-auto">
                <div class="animate-pulse">
                  <div class="mb-8 mt-0 aspect-video w-full rounded bg-gray-300"></div>
                  <div class="space-y-4">
                    <div class="h-8 bg-gray-300 rounded w-3/4"></div>
                    <div class="h-4 bg-gray-300 rounded w-full"></div>
                    <div class="h-4 bg-gray-300 rounded w-5/6"></div>
                    <div class="h-4 bg-gray-300 rounded w-4/5"></div>
                  </div>
                </div>
              </article>
            </div>

            <!-- Sidebar Loading -->
            <div class="order-1 md:order-none md:col-span-5 lg:col-span-4">
              <div class="md:sticky md:top-20">
                <aside>
                  <div class="animate-pulse space-y-8">
                    <div class="bg-gray-200 rounded-lg p-6">
                      <div class="space-y-3">
                        <div class="h-4 bg-gray-300 rounded w-1/2"></div>
                        <div class="h-4 bg-gray-300 rounded w-3/4"></div>
                        <div class="h-4 bg-gray-300 rounded w-2/3"></div>
                      </div>
                    </div>
                  </div>
                </aside>
              </div>
            </div>
          </div>
        </div>
      </section>
    </div>

    <!-- Content State -->
    <div v-else-if="serviceData">
      <section class="pb-8">
        <ServiceBreadcrumb
          :serviceName="serviceData.title"
          :title="serviceData.title"
          :category="serviceData.category"
        />

        <div class="container service-page-container">
          <div class="mt-8 grid gap-12 md:grid-cols-12 md:gap-8">
            <div class="order-2 md:order-none md:col-span-7 md:col-start-1 lg:col-span-8">
              <article class="prose dark:prose-invert mx-auto">
                <div>
                  <img
                    :src="serviceData.heroImage.src"
                    :alt="serviceData.heroImage.alt"
                    class="mb-8 mt-0 aspect-video w-full rounded object-cover"
                  />
                </div>

                <ServiceContent :service="serviceData" />
              </article>
            </div>

            <div class="order-1 md:order-none md:col-span-5 lg:col-span-4">
              <div class="md:sticky md:top-20">
                <aside id="service-page-aside">
                  <ServiceTreatmentDetails
                    :duration="serviceData.treatmentDetails.duration"
                    :recovery="serviceData.treatmentDetails.recovery"
                    :deliveryModes="serviceData.deliveryModes"
                    :isComingSoon="serviceData.isComingSoon"
                  />

                  <ServiceContact class="mt-8" />
                </aside>
              </div>
            </div>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import ServiceBreadcrumb from './ServiceBreadcrumb.vue';
import ServiceContent from './ServiceContent.vue';
import ServiceTreatmentDetails from './ServiceTreatmentDetails.vue';
import ServiceContact from './ServiceContact.vue';

interface Service {
  id: string;
  title: string;
  slug: string;
  description: string;
  detailed_description?: string;
  technologies?: string[];
  features?: string[];
  delivery_modes?: string[];
  image?: string;
  status: string;
  category_id?: number;
}

interface ServicePageData {
  id: string;
  title: string;
  slug: string;
  description: string;
  detailed_description?: string;
  technologies?: string[];
  features?: string[];
  heroImage: {
    src: string;
    alt: string;
  };
  treatmentDetails: {
    duration: string;
    recovery: string;
  };
  deliveryModes: string[];
  category?: string;
  isComingSoon: boolean;
}

// Reactive state
const service = ref<Service | null>(null);
const isLoading = ref(false);
const error = ref<string | null>(null);
const categories = ref<{ id: number; name: string }[]>([]);

// Transform service data to match the expected structure
const serviceData = computed((): ServicePageData | null => {
  if (!service.value) return null;
  
  // Find category name from category_id
  const categoryName = categories.value.find(cat => cat.id === service.value?.category_id)?.name;
  
  return {
    id: service.value.id,
    title: service.value.title,
    slug: service.value.slug,
    description: service.value.description,
    detailed_description: service.value.detailed_description,
    technologies: service.value.technologies,
    features: service.value.features,
    heroImage: {
      src: service.value.image || `https://placehold.co/1200x600/2563eb/ffffff/png?text=${encodeURIComponent(service.value.title)}`,
      alt: `${service.value.title} service at International Center`,
    },
    treatmentDetails: {
      duration: '45-90 minutes',
      recovery: 'Minimal to no downtime',
    },
    deliveryModes: parseDeliveryModes(service.value.slug),
    category: categoryName,
    isComingSoon: false,
  };
});

// Parse delivery modes based on service type (from original logic)
const parseDeliveryModes = (slug: string): string[] => {
  const modes: string[] = [];

  // Mobile services (can be performed at patient location)
  if (
    [
      'prp-therapy',
      'exosome-therapy',
      'peptide-therapy',
      'iv-therapy',
      'wellness',
      'immunizations',
      'telehealth',
      'annual-wellness',
      'chronic-care',
      'physical-exams',
      'immune-support',
    ].includes(slug)
  ) {
    modes.push('mobile');
  }

  // Outpatient services (most services are outpatient except complex procedures)
  if (!['stem-cell'].includes(slug)) {
    modes.push('outpatient');
  }

  // Inpatient services (requiring facility stay or complex procedures)
  if (['stem-cell', 'diagnostics', 'longevity'].includes(slug)) {
    modes.push('inpatient');
  }

  return modes.length > 0 ? modes : ['outpatient'];
};

// Get slug from current URL
const getSlugFromUrl = (): string => {
  if (typeof window === 'undefined') return '';
  const pathParts = window.location.pathname.split('/');
  return pathParts[pathParts.length - 1] || '';
};

// Client-side data loading following home page pattern
onMounted(async () => {
  try {
    isLoading.value = true;
    error.value = null;

    const slug = getSlugFromUrl();
    if (!slug) {
      throw new Error('Service slug not found');
    }

    console.log(`🔍 [DynamicServicePage] Loading service: ${slug}`);

    // Dynamic import for client-side code splitting (like RecentContentCycler)
    const { servicesClient } = await import('../../lib/clients');
    
    // Fetch both service data and categories in parallel from REST API
    const [serviceResponse, categoriesResponse] = await Promise.all([
      servicesClient.getServiceBySlug(slug),
      servicesClient.getServiceCategories()
    ]);
    
    // Handle REST API response format
    if (!serviceResponse.success) {
      throw new Error(serviceResponse.message || `Service not found: ${slug}`);
    }
    
    if (!categoriesResponse.success) {
      throw new Error(categoriesResponse.message || 'Failed to load categories');
    }
    
    service.value = serviceResponse.data;
    categories.value = categoriesResponse.data;
    console.log('✅ [DynamicServicePage] Service loaded:', serviceResponse.data);
    console.log('✅ [DynamicServicePage] Categories loaded:', categoriesResponse.data);
  } catch (err: any) {
    console.error('❌ [DynamicServicePage] Failed to load service:', err.message);
    error.value = err.message || 'Failed to load service data';
  } finally {
    isLoading.value = false;
  }
});
</script>

<style scoped>
.service-page-container {
  overflow: visible !important;
}

.prose {
  max-width: none;
}
</style>