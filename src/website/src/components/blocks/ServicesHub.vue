<template>
  <div>
    <!-- Error State -->
    <div v-if="servicesError" class="text-center py-12">
      <div class="max-w-md mx-auto">
        <h3 class="text-lg font-semibold text-gray-900 mb-2">Services Temporarily Unavailable</h3>
        <p class="text-gray-600 mb-4">{{ servicesError }}</p>
        <a
          href="/"
          class="inline-block px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
        >
          Return Home
        </a>
      </div>
    </div>

    <!-- Loading State -->
    <div v-else-if="servicesLoading || categoriesLoading">
      <!-- Filter Section Loading -->
      <section class="pt-8 lg:pt-12">
        <div class="container mx-auto px-4">
          <div class="animate-pulse">
            <div class="bg-gray-50 dark:bg-gray-800/50 rounded py-3 px-6">
              <div class="flex flex-wrap gap-2 justify-center">
                <div class="h-10 bg-gray-300 rounded w-20"></div>
                <div class="h-10 bg-gray-300 rounded w-24"></div>
                <div class="h-10 bg-gray-300 rounded w-28"></div>
                <div class="h-10 bg-gray-300 rounded w-22"></div>
              </div>
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
          <ServicesFilter
            :enable-delivery-mode-filter="true"
            :enable-category-filter="false"
            @filtered-services="handleFilteredServices"
          />
        </div>
      </section>

      <!-- Services Categories -->
      <section class="pt-6 lg:pt-10 pb-8 lg:pb-12">
        <div class="container mx-auto px-4">
          <div v-if="displayedServices.length === 0" class="text-center py-12">
            <div class="max-w-md mx-auto">
              <h3 class="text-lg font-semibold text-gray-900 mb-2">
                No Services Found
              </h3>
              <p class="text-gray-600 text-sm">
                No services match your current filter criteria.
              </p>
            </div>
          </div>

          <div v-else class="space-y-12 lg:space-y-16">
            <div v-for="category in groupedServices" :key="category.category_id" class="service-category">
              <div class="mb-8">
                <h2 class="text-3xl font-semibold lg:text-4xl text-gray-900 dark:text-white mb-2">
                  {{ category.name }}
                </h2>
                <div class="w-12 h-px bg-black dark:bg-white"></div>
              </div>

              <div class="grid gap-4 sm:gap-6 md:grid-cols-2 lg:grid-cols-3" role="list">
                <RouterLink
                  v-for="service in category.services"
                  :key="service.service_id"
                  :to="`/services/${service.slug}`"
                  class="block group service-card rounded-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
                  :aria-label="`${service.title} - ${getServiceStatus(service)}`"
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
                        {{ service.title }}
                      </h3>
                      <Badge
                        v-if="service.publishing_status === 'published'"
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
                      <!-- Delivery Mode -->
                      <div class="flex items-center text-sm text-gray-600 dark:text-gray-400">
                        <span
                          class="w-3 h-px bg-black dark:bg-white mr-3 mt-0.5 flex-shrink-0"
                        ></span>
                        Type: {{ formatDeliveryMode(service.delivery_mode) }}
                      </div>

                      <!-- Order info if available -->
                      <div v-if="service.order_number" class="flex items-center text-sm text-gray-600 dark:text-gray-400">
                        <span
                          class="w-3 h-px bg-black dark:bg-white mr-3 mt-0.5 flex-shrink-0"
                        ></span>
                        Order: #{{ service.order_number }}
                      </div>
                    </div>

                    <!-- Service Status Indicator -->
                    <div class="mt-6 pt-4 border-t border-gray-100 dark:border-gray-800">
                      <div class="flex items-center justify-between">
                        <span
                          class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wide"
                        >
                          {{ getServiceStatus(service) }}
                        </span>
                        <div class="flex space-x-1" aria-label="Service delivery mode">
                          <div
                            :class="['w-2 h-2 rounded-full', getDeliveryModeColor(service.delivery_mode)]"
                            :title="formatDeliveryMode(service.delivery_mode)"
                            aria-hidden="true"
                          ></div>
                          <span class="sr-only">
                            Service type: {{ formatDeliveryMode(service.delivery_mode) }}
                          </span>
                        </div>
                      </div>
                    </div>
                  </div>
                </RouterLink>
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
import { RouterLink } from 'vue-router';
import { Badge } from '../vue-ui';
import ServicesFilter from '../ServicesFilter.vue';
import { useServices, useServiceCategories } from '../../composables/useServices';
import type { Service } from '../../lib/clients/services/types';

// Reactive data from stores
const { services, loading: servicesLoading, error: servicesError } = useServices();
const { categories, loading: categoriesLoading } = useServiceCategories();

// Filtered services state
const displayedServices = ref<Service[]>([]);

// Grouped services by category for display
const groupedServices = computed(() => {
  const grouped = new Map();
  
  displayedServices.value.forEach(service => {
    const category = categories.value.find(cat => cat.category_id === service.category_id);
    if (!category) return;
    
    if (!grouped.has(category.category_id)) {
      grouped.set(category.category_id, {
        category_id: category.category_id,
        name: category.name,
        slug: category.slug,
        services: []
      });
    }
    
    grouped.get(category.category_id).services.push(service);
  });
  
  // Sort by category order
  return Array.from(grouped.values()).sort((a, b) => {
    const categoryA = categories.value.find(cat => cat.category_id === a.category_id);
    const categoryB = categories.value.find(cat => cat.category_id === b.category_id);
    return (categoryA?.order_number || 0) - (categoryB?.order_number || 0);
  });
});

// Handle filtered services from ServicesFilter component
const handleFilteredServices = (filteredServices: Service[]) => {
  displayedServices.value = filteredServices;
};

// Utility functions
const formatDeliveryMode = (deliveryMode: string): string => {
  const modes = {
    mobile_service: 'Mobile Service',
    outpatient_service: 'Outpatient Service',
    inpatient_service: 'Inpatient Service',
  };
  return modes[deliveryMode as keyof typeof modes] || deliveryMode;
};

const getServiceStatus = (service: Service): string => {
  return service.publishing_status === 'published' ? 'Available Now' : 'Coming Soon';
};

const getDeliveryModeColor = (deliveryMode: string): string => {
  const colors = {
    mobile_service: 'bg-blue-400',
    outpatient_service: 'bg-green-400',
    inpatient_service: 'bg-purple-400',
  };
  return colors[deliveryMode as keyof typeof colors] || 'bg-gray-400';
};

// Initialize data on mount
onMounted(async () => {
  // Initial load - show all services
  displayedServices.value = services.value.filter(service => service.publishing_status === 'published');
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
