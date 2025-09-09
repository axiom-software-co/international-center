<template>
  <div class="bg-gray-50 dark:bg-gray-800/50 rounded py-3 px-6">
    <div class="flex flex-wrap items-center gap-2">
      <span class="text-sm text-gray-600 dark:text-gray-400 mr-2">{{ label }}:</span>

      <!-- Mobile Dropdown - Hidden on md and up -->
      <div class="md:hidden w-full">
        <Select v-model="selectedFilter" @update:modelValue="handleFilterChange">
          <SelectTrigger class="w-full">
            <SelectValue :placeholder="placeholder" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem
              v-for="option in filterOptions"
              :key="option.value"
              :value="option.value"
            >
              {{ option.label }}
              <span v-if="option.count !== undefined" class="ml-1 text-gray-500">({{ option.count }})</span>
            </SelectItem>
          </SelectContent>
        </Select>
      </div>

      <!-- Desktop Buttons - Hidden on mobile -->
      <div class="hidden md:flex md:flex-wrap md:items-center md:gap-2">
        <button
          v-for="option in filterOptions"
          :key="option.value"
          :class="[
            'px-3 py-1 text-sm font-medium rounded transition-colors',
            isActive(option.value)
              ? 'bg-blue-500 text-white border-blue-500'
              : 'border border-gray-300 bg-white text-gray-900 md:hover:bg-gray-50',
          ]"
          @click="handleFilterChange(option.value)"
        >
          {{ option.label }}
          <span v-if="option.count !== undefined" class="ml-1 opacity-75">({{ option.count }})</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/vue-ui';
import type { FilterOption } from '../../composables/filtering/useFiltering';

interface GenericFilterProps {
  filterName: string;
  label?: string;
  placeholder?: string;
  options: FilterOption[];
  multiple?: boolean;
  modelValue?: string | string[];
}

interface GenericFilterEmits {
  (e: 'update:modelValue', value: string | string[]): void;
  (e: 'filter-change', filterName: string, value: string | string[]): void;
}

const props = withDefaults(defineProps<GenericFilterProps>(), {
  label: 'Filter by',
  placeholder: 'Select filter',
  multiple: false,
  modelValue: '',
});

const emit = defineEmits<GenericFilterEmits>();

// Computed filter options with all option
const filterOptions = computed(() => {
  const allOption: FilterOption = { 
    value: '', 
    label: props.multiple ? 'All' : `All ${props.label}` 
  };
  return [allOption, ...props.options];
});

// Internal state for selected filter
const selectedFilter = ref<string | string[]>(
  props.multiple ? [] : (props.modelValue || '')
);

// Watch for external changes to modelValue
watch(() => props.modelValue, (newValue) => {
  selectedFilter.value = newValue || (props.multiple ? [] : '');
}, { immediate: true });

// Check if a filter value is active
const isActive = (value: string): boolean => {
  if (props.multiple) {
    return Array.isArray(selectedFilter.value) 
      ? selectedFilter.value.includes(value)
      : false;
  }
  return selectedFilter.value === value;
};

// Handle filter changes
const handleFilterChange = (value: string): void => {
  if (props.multiple) {
    // Handle multiple selection
    const currentValues = Array.isArray(selectedFilter.value) ? selectedFilter.value : [];
    
    if (value === '') {
      // Clear all if "All" selected
      selectedFilter.value = [];
    } else {
      // Toggle the selected value
      const index = currentValues.indexOf(value);
      if (index >= 0) {
        selectedFilter.value = currentValues.filter(v => v !== value);
      } else {
        selectedFilter.value = [...currentValues, value];
      }
    }
  } else {
    // Handle single selection
    selectedFilter.value = value;
  }

  // Emit changes
  emit('update:modelValue', selectedFilter.value);
  emit('filter-change', props.filterName, selectedFilter.value);
};
</script>