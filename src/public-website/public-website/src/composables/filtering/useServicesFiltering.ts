// Services-specific filtering composable
// Integrates services data with generic filtering system

import { ref, computed, watch, type Ref } from 'vue';
import { useServices, useServiceCategories } from '../useServices';
import { useFiltering, useCategoryFiltering, type FilterConfig, type FilterOption } from './useFiltering';
import type { Service, ServiceCategory } from '../../lib/clients/services/types';

export interface UseServicesFilteringOptions {
  enableCategoryFilter?: boolean;
  enableDeliveryModeFilter?: boolean;
  includeCounts?: boolean;
}

export interface UseServicesFilteringResult {
  // Filtered services data
  filteredServices: Ref<Service[]>;
  
  // Filter state
  filters: Ref<{ [key: string]: string | string[] }>;
  hasActiveFilters: Ref<boolean>;
  
  // Filter configurations for UI
  categoryFilterConfig: Ref<FilterConfig | null>;
  deliveryModeFilterConfig: Ref<FilterConfig | null>;
  
  // Filter actions
  setCategoryFilter: (categoryId: string) => void;
  setDeliveryModeFilter: (mode: string) => void;
  clearAllFilters: () => void;
  
  // Utility methods
  isFilterActive: (filterName: string, value?: string) => boolean;
}

// Delivery mode labels mapping
const DELIVERY_MODE_LABELS = {
  mobile_service: 'Mobile Service',
  outpatient_service: 'Outpatient Service', 
  inpatient_service: 'Inpatient Service',
} as const;

export function useServicesFiltering(
  servicesOptions: Parameters<typeof useServices>[0] = {},
  filterOptions: UseServicesFilteringOptions = {}
): UseServicesFilteringResult {

  const {
    enableCategoryFilter = true,
    enableDeliveryModeFilter = true,
    includeCounts = true,
  } = filterOptions;

  // Load services and categories data
  const { services, loading: servicesLoading } = useServices(servicesOptions);
  const { categories, loading: categoriesLoading } = useServiceCategories();

  // Create filter configurations
  const filterConfigs: FilterConfig[] = [];

  // Category filter configuration
  const categoryFilterConfig = computed(() => {
    if (!enableCategoryFilter || categoriesLoading.value) return null;

    const categoryOptions: FilterOption[] = categories.value.map(category => ({
      value: category.category_id,
      label: category.name,
      count: includeCounts ? countServicesInCategory(services.value, category.category_id) : undefined,
    }));

    return {
      name: 'category',
      label: 'Category',
      options: categoryOptions,
      multiple: false,
      searchable: true,
    };
  });

  // Delivery mode filter configuration  
  const deliveryModeFilterConfig = computed(() => {
    if (!enableDeliveryModeFilter || servicesLoading.value) return null;

    const deliveryModeOptions: FilterOption[] = Object.entries(DELIVERY_MODE_LABELS).map(([mode, label]) => ({
      value: mode,
      label,
      count: includeCounts ? countServicesByDeliveryMode(services.value, mode as keyof typeof DELIVERY_MODE_LABELS) : undefined,
    }));

    return {
      name: 'delivery_mode',
      label: 'Delivery Mode',
      options: deliveryModeOptions,
      multiple: false,
      searchable: false,
    };
  });

  // Build dynamic filter configs
  watch([categoryFilterConfig, deliveryModeFilterConfig], ([categoryConfig, deliveryModeConfig]) => {
    filterConfigs.length = 0; // Clear existing configs
    
    if (categoryConfig) {
      filterConfigs.push(categoryConfig);
    }
    if (deliveryModeConfig) {
      filterConfigs.push(deliveryModeConfig);
    }
  }, { immediate: true });

  // Initialize filtering system
  const filteringResult = useFiltering(services, filterConfigs);

  // Convenient filter setters
  const setCategoryFilter = (categoryId: string): void => {
    filteringResult.setFilter('category', categoryId);
  };

  const setDeliveryModeFilter = (mode: string): void => {
    filteringResult.setFilter('delivery_mode', mode);
  };

  return {
    filteredServices: filteringResult.filteredItems,
    filters: filteringResult.filters,
    hasActiveFilters: filteringResult.hasActiveFilters,
    categoryFilterConfig,
    deliveryModeFilterConfig,
    setCategoryFilter,
    setDeliveryModeFilter,
    clearAllFilters: filteringResult.clearAllFilters,
    isFilterActive: filteringResult.isFilterActive,
  };
}

// Helper functions
function countServicesInCategory(services: Service[], categoryId: string): number {
  return services.filter(service => service.category_id === categoryId).length;
}

function countServicesByDeliveryMode(services: Service[], deliveryMode: string): number {
  return services.filter(service => service.delivery_mode === deliveryMode).length;
}

// Category-only filtering composable for simpler use cases
export function useServicesCategoryFiltering(
  servicesOptions: Parameters<typeof useServices>[0] = {}
): UseServicesFilteringResult {
  return useServicesFiltering(servicesOptions, {
    enableCategoryFilter: true,
    enableDeliveryModeFilter: false,
    includeCounts: true,
  });
}

// Delivery mode-only filtering composable
export function useServicesDeliveryModeFiltering(
  servicesOptions: Parameters<typeof useServices>[0] = {}
): UseServicesFilteringResult {
  return useServicesFiltering(servicesOptions, {
    enableCategoryFilter: false,
    enableDeliveryModeFilter: true,
    includeCounts: true,
  });
}