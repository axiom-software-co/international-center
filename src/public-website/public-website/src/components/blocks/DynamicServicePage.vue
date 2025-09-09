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
import { ref, computed } from 'vue';
import { useService, useServiceCategories } from '../../composables/useServices';
import { getServiceSlugFromUrl } from '../../lib/utils/url';
import { parseServiceDeliveryModes, generateHeroImageUrl, generateImageAlt } from '../../lib/utils/content';
import ServiceBreadcrumb from './ServiceBreadcrumb.vue';
import ServiceContent from './ServiceContent.vue';
import ServiceTreatmentDetails from './ServiceTreatmentDetails.vue';
import ServiceContact from './ServiceContact.vue';
import type { Service, ServiceCategory } from '../../lib/clients';

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

// Use composables for data fetching
const currentSlug = ref(getServiceSlugFromUrl());
const { service, loading: serviceLoading, error: serviceError } = useService(currentSlug);
const { categories, loading: categoriesLoading, error: categoriesError } = useServiceCategories();

// Computed loading and error states
const isLoading = computed(() => serviceLoading.value || categoriesLoading.value);
const error = computed(() => serviceError.value || categoriesError.value);

// Transform service data to match the expected structure
const serviceData = computed((): ServicePageData | null => {
  if (!service.value) return null;
  
  // Find category name from category_id
  const categoryName = categories.value.find((cat: ServiceCategory) => cat.category_id === service.value?.category_id)?.name;
  
  return {
    id: service.value.service_id,
    title: service.value.title,
    slug: service.value.slug,
    description: service.value.description,
    detailed_description: service.value.content,
    technologies: [],
    features: [],
    heroImage: {
      src: generateHeroImageUrl(service.value.image_url, service.value.title, 'service'),
      alt: generateImageAlt(service.value.title, 'service'),
    },
    treatmentDetails: {
      duration: '45-90 minutes',
      recovery: 'Minimal to no downtime',
    },
    deliveryModes: parseServiceDeliveryModes(service.value.slug),
    category: categoryName,
    isComingSoon: false,
  };
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