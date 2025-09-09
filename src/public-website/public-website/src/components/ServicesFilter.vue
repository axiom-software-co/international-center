<template>
  <!-- Delivery Mode Filter -->
  <GenericFilter
    v-if="deliveryModeConfig"
    filter-name="delivery_mode"
    label="Filter by"
    placeholder="Select delivery mode"
    :options="deliveryModeConfig.options"
    :model-value="currentDeliveryMode"
    @filter-change="handleDeliveryModeChange"
  />
  
  <!-- Category Filter (if enabled) -->
  <GenericFilter
    v-if="enableCategoryFilter && categoryConfig"
    filter-name="category"
    label="Category"
    placeholder="Select category"
    :options="categoryConfig.options"
    :model-value="currentCategory"
    @filter-change="handleCategoryChange"
  />
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import GenericFilter from './base/GenericFilter.vue';
import { useServicesFiltering, useServicesDeliveryModeFiltering } from '../composables/filtering/useServicesFiltering';

interface ServicesFilterProps {
  enableCategoryFilter?: boolean;
  enableDeliveryModeFilter?: boolean;
  onFilterChange?: (filters: { delivery_mode?: string; category?: string }) => void;
}

interface ServicesFilterEmits {
  (e: 'filtered-services', services: any[]): void;
  (e: 'filter-change', filters: { delivery_mode?: string; category?: string }): void;
}

const props = withDefaults(defineProps<ServicesFilterProps>(), {
  enableCategoryFilter: false,
  enableDeliveryModeFilter: true,
});

const emit = defineEmits<ServicesFilterEmits>();

// Use the services filtering composable
const {
  filteredServices,
  deliveryModeFilterConfig,
  categoryFilterConfig,
  filters,
  setDeliveryModeFilter,
  setCategoryFilter,
} = useServicesFiltering({}, {
  enableCategoryFilter: props.enableCategoryFilter,
  enableDeliveryModeFilter: props.enableDeliveryModeFilter,
  includeCounts: true,
});

// Current filter values
const currentDeliveryMode = computed(() => {
  const value = filters.value.delivery_mode;
  return Array.isArray(value) ? value[0] || '' : value || '';
});

const currentCategory = computed(() => {
  const value = filters.value.category;
  return Array.isArray(value) ? value[0] || '' : value || '';
});

// Filter configurations for UI
const deliveryModeConfig = computed(() => deliveryModeFilterConfig.value);
const categoryConfig = computed(() => categoryFilterConfig.value);

// Handle filter changes
const handleDeliveryModeChange = (filterName: string, value: string | string[]) => {
  const filterValue = Array.isArray(value) ? value[0] || '' : value || '';
  setDeliveryModeFilter(filterValue);
  emitFilterChange();
};

const handleCategoryChange = (filterName: string, value: string | string[]) => {
  const filterValue = Array.isArray(value) ? value[0] || '' : value || '';
  setCategoryFilter(filterValue);
  emitFilterChange();
};

// Emit filter changes
const emitFilterChange = () => {
  const filterValues = {
    delivery_mode: currentDeliveryMode.value || undefined,
    category: currentCategory.value || undefined,
  };
  
  // Remove empty values
  Object.keys(filterValues).forEach(key => {
    if (!filterValues[key as keyof typeof filterValues]) {
      delete filterValues[key as keyof typeof filterValues];
    }
  });

  emit('filter-change', filterValues);
  
  if (props.onFilterChange) {
    props.onFilterChange(filterValues);
  }
};

// Watch filtered services and emit them
watch(filteredServices, (services) => {
  emit('filtered-services', services);
}, { immediate: true });
</script>
