<template>
  <div>
    <!-- Error State -->
    <div v-if="error" class="text-center py-12">
      <div class="max-w-md mx-auto">
        <h3 class="text-lg font-semibold text-gray-900 mb-2">Services Temporarily Unavailable</h3>
        <p class="text-gray-600 mb-4">
          We're experiencing technical difficulties. Please try again later.
        </p>
        <a
          href="/"
          class="inline-block px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
        >
          Return Home
        </a>
      </div>
    </div>

    <!-- Loading State -->
    <div v-else-if="isLoading">
      <!-- Filter Section Loading -->
      <section class="pt-8 lg:pt-12">
        <div class="container mx-auto px-4">
          <div class="animate-pulse">
            <div class="flex flex-wrap gap-2 justify-center">
              <div class="h-10 bg-gray-300 rounded w-20"></div>
              <div class="h-10 bg-gray-300 rounded w-24"></div>
              <div class="h-10 bg-gray-300 rounded w-28"></div>
              <div class="h-10 bg-gray-300 rounded w-22"></div>
            </div>
          </div>
        </div>
      </section>

      <!-- Services Categories Loading -->
      <section class="pt-6 lg:pt-10 pb-8 lg:pb-12">
        <div class="container mx-auto px-4">
          <div class="space-y-12 lg:space-y-16">
            <div v-for="n in 3" :key="n" class="service-category">
              <div class="mb-8 animate-pulse">
                <div class="h-8 bg-gray-300 rounded w-64 mb-2"></div>
                <div class="w-12 h-px bg-gray-300"></div>
              </div>

              <div class="grid gap-4 sm:gap-6 md:grid-cols-2 lg:grid-cols-3">
                <div v-for="i in 6" :key="i" class="animate-pulse">
                  <div class="bg-gray-200 border border-gray-300 rounded-sm p-6 h-full">
                    <div class="flex items-start justify-between mb-4">
                      <div class="h-6 bg-gray-300 rounded w-3/4"></div>
                      <div class="h-6 bg-gray-300 rounded w-16"></div>
                    </div>
                    <div class="space-y-2 mb-6">
                      <div class="h-4 bg-gray-300 rounded w-full"></div>
                      <div class="h-4 bg-gray-300 rounded w-5/6"></div>
                    </div>
                    <div class="space-y-3">
                      <div class="h-4 bg-gray-300 rounded w-2/3"></div>
                      <div class="h-4 bg-gray-300 rounded w-3/4"></div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>
    </div>

    <!-- Content State -->
    <div v-else>
      <!-- Filter Section -->
      <section class="pt-8 lg:pt-12">
        <div class="container mx-auto px-4">
          <ServicesFilter :filter-options="filterOptions" />
        </div>
      </section>

      <!-- Services Categories -->
      <section class="pt-6 lg:pt-10 pb-8 lg:pb-12">
        <div class="container mx-auto px-4">
          <div v-if="serviceCategories.length === 0" class="text-center py-12">
            <div class="max-w-md mx-auto">
              <h3 class="text-lg font-semibold text-gray-900 mb-2">
                Services Temporarily Unavailable
              </h3>
              <p class="text-gray-600 text-sm">
                We're unable to load service information at the moment.
              </p>
            </div>
          </div>

          <div v-else class="space-y-12 lg:space-y-16">
            <div v-for="category in serviceCategories" :key="category.id" class="service-category">
              <div class="mb-8">
                <h2 class="text-3xl font-semibold lg:text-4xl text-gray-900 dark:text-white mb-2">
                  {{ category.title }}
                </h2>
                <div class="w-12 h-px bg-black dark:bg-white"></div>
              </div>

              <div class="grid gap-4 sm:gap-6 md:grid-cols-2 lg:grid-cols-3" role="list">
                <a
                  v-for="service in category.services"
                  :key="service.href"
                  :href="service.href"
                  class="block group service-card rounded-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
                  :data-available="service.available"
                  v-bind="getServiceDataAttributes(service)"
                  :aria-label="`${service.name} - ${service.available ? 'Available now' : 'Coming soon'}`"
                  role="listitem"
                >
                  <div
                    class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-sm p-6 h-full transition-all duration-200 hover:shadow-sm group-focus:shadow-sm"
                  >
                    <!-- Service Header -->
                    <div class="flex items-start justify-between mb-4">
                      <h3
                        class="text-lg font-semibold text-gray-900 dark:text-white group-hover:text-gray-700 dark:group-hover:text-gray-300 transition-colors leading-tight"
                      >
                        {{ service.name }}
                      </h3>
                      <Badge
                        variant="outline"
                        class="ml-2 px-2 py-0.5 text-xs font-medium border border-green-300 bg-green-50 text-green-700 rounded shrink-0"
                      >
                        Available
                      </Badge>
                    </div>

                    <!-- Service Description -->
                    <p class="text-gray-600 dark:text-gray-400 mb-6 text-sm leading-relaxed">
                      {{ service.description }}
                    </p>

                    <!-- Service Details -->
                    <div class="space-y-3">
                      <!-- Duration -->
                      <div class="flex items-center text-sm text-gray-600 dark:text-gray-400">
                        <span
                          class="w-3 h-px bg-black dark:bg-white mr-3 mt-0.5 flex-shrink-0"
                        ></span>
                        Duration: {{ service.duration }}
                      </div>

                      <!-- Delivery Modes -->
                      <div class="flex items-center text-sm text-gray-600 dark:text-gray-400">
                        <span
                          class="w-3 h-px bg-black dark:bg-white mr-3 mt-0.5 flex-shrink-0"
                        ></span>
                        Available: {{ getDeliveryModes(service) }}
                      </div>
                    </div>

                    <!-- Service Status Indicator -->
                    <div class="mt-6 pt-4 border-t border-gray-100 dark:border-gray-800">
                      <div class="flex items-center justify-between">
                        <span
                          class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wide"
                        >
                          {{ service.available ? 'Available Now' : 'Coming Soon' }}
                        </span>
                        <div class="flex space-x-1" aria-label="Service delivery modes">
                          <div
                            v-for="(mode, index) in service.delivery_modes"
                            :key="mode"
                            :class="['w-2 h-2 rounded-full', getDeliveryModeColor(index)]"
                            :title="`${mode.charAt(0).toUpperCase() + mode.slice(1)} Service`"
                            aria-hidden="true"
                          ></div>
                          <span class="sr-only">
                            Service types: {{ getDeliveryModes(service) }}
                          </span>
                        </div>
                      </div>
                    </div>
                  </div>
                </a>
              </div>
            </div>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { Badge } from '../vue-ui';
import ServicesFilter from '../ServicesFilter.vue';

const serviceCategories = ref<any[]>([]);
const isLoading = ref(false);
const error = ref<string | null>(null);
const filterOptions = ref<Array<{ value: string; label: string }>>([
  { value: 'all', label: 'Show All' },
]);

const getDeliveryModes = (service: any): string => {
  if (Array.isArray(service.delivery_modes) && service.delivery_modes.length > 0) {
    return service.delivery_modes
      .map((mode: string) => mode.charAt(0).toUpperCase() + mode.slice(1))
      .join(', ');
  }
  return 'Contact for details';
};

const getServiceDataAttributes = (service: any): Record<string, any> => {
  const attributes: Record<string, any> = {};

  if (Array.isArray(service.delivery_modes)) {
    service.delivery_modes.forEach((mode: string) => {
      attributes[`data-${mode}`] = 'true';
    });
  }

  return attributes;
};

const getDeliveryModeColor = (index: number): string => {
  const colors = [
    'bg-blue-400',
    'bg-green-400',
    'bg-purple-400',
    'bg-orange-400',
    'bg-pink-400',
    'bg-indigo-400',
    'bg-teal-400',
    'bg-red-400',
  ];
  return colors[index % colors.length];
};

onMounted(async () => {
  try {
    isLoading.value = true;
    error.value = null;

    const { loadServicesPageData } = await import('../../lib/navigation-data');
    const servicesData = await loadServicesPageData();
    console.log('🔍 [ServicesHub] Client-side data loaded:', servicesData);

    // Update reactive state with new data
    serviceCategories.value = servicesData.serviceCategories;

    // Calculate available filter options from actual services data
    const availableFilters = new Set<string>();

    // Always include 'Show All' option
    const dynamicFilterOptions = [{ value: 'all', label: 'Show All' }];

    // Analyze all services to determine which delivery modes are available
    servicesData.serviceCategories.forEach(category => {
      category.services.forEach((service: any) => {
        if (Array.isArray(service.delivery_modes)) {
          service.delivery_modes.forEach((mode: string) => {
            availableFilters.add(mode);
          });
        }
      });
    });

    // Add filter options based on what's actually available in the data
    Array.from(availableFilters)
      .sort()
      .forEach(mode => {
        const label = `${mode.charAt(0).toUpperCase() + mode.slice(1)} Services`;
        dynamicFilterOptions.push({ value: mode, label });
      });

    filterOptions.value = dynamicFilterOptions;
  } catch (err: any) {
    console.error('❌ [ServicesHub] Client-side data loading failed:', err);
    error.value = 'Failed to load services. Please try again later.';
    serviceCategories.value = [];
    // Reset to default filter options on error
    filterOptions.value = [{ value: 'all', label: 'Show All' }];
  } finally {
    isLoading.value = false;
  }
});
</script>

<style scoped>
.filter-btn.active {
  background-color: #3b82f6;
  color: white;
  border-color: #3b82f6;
}

@media (min-width: 768px) {
  .filter-btn.active:hover {
    background-color: #2563eb;
  }
}

.service-card {
  transition:
    opacity 0.3s ease,
    transform 0.2s ease,
    box-shadow 0.2s ease;
}

.service-card:hover {
  transform: translateY(-1px);
}

.service-card.hidden {
  display: none;
}

/* Service delivery mode indicators */
.service-card[data-mobile='true'] .delivery-indicator::before {
  content: '';
  position: absolute;
  top: -2px;
  left: -2px;
  right: -2px;
  bottom: -2px;
  border: 2px solid #60a5fa;
  border-radius: 2px;
  opacity: 0;
  transition: opacity 0.2s ease;
}

.service-card[data-outpatient='true'] .delivery-indicator::after {
  content: '';
  position: absolute;
  top: -2px;
  left: -2px;
  right: -2px;
  bottom: -2px;
  border: 2px solid #34d399;
  border-radius: 2px;
  opacity: 0;
  transition: opacity 0.2s ease;
}

.service-card[data-inpatient='true']:hover .delivery-indicator::before {
  border-color: #a78bfa;
}

/* Enhanced focus states for accessibility */
.service-card:focus-visible {
  outline: 2px solid #3b82f6;
  outline-offset: 2px;
}

/* Smooth category transitions */
.service-category {
  transition: opacity 0.3s ease;
}

/* Line clamp fallback for older browsers */
@supports not (-webkit-line-clamp: 2) {
  .line-clamp-2 {
    overflow: hidden;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
  }
}

/* Screen reader only utility */
.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}
</style>
