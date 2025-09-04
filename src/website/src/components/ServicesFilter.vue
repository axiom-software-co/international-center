<template>
  <div class="bg-gray-50 dark:bg-gray-800/50 rounded py-3 px-6">
    <div class="flex flex-wrap items-center gap-2">
      <span class="text-sm text-gray-600 dark:text-gray-400 mr-2">Filter by:</span>

      <!-- Mobile Dropdown - Hidden on md and up -->
      <div class="md:hidden w-full">
        <Select v-model="activeFilter" @update:modelValue="handleFilterChange">
          <SelectTrigger class="w-full">
            <SelectValue placeholder="Select filter" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem
              v-for="option in props.filterOptions"
              :key="option.value"
              :value="option.value"
            >
              {{ option.label }}
            </SelectItem>
          </SelectContent>
        </Select>
      </div>

      <!-- Desktop Buttons - Hidden on mobile -->
      <div class="hidden md:flex md:flex-wrap md:items-center md:gap-2">
        <button
          v-for="option in props.filterOptions"
          :key="option.value"
          :class="[
            'filter-btn px-3 py-1 text-sm font-medium rounded transition-colors',
            activeFilter === option.value
              ? 'active bg-blue-500 text-white border-blue-500'
              : 'border border-gray-300 bg-white text-gray-900 md:hover:bg-gray-50',
          ]"
          :data-filter="option.value"
          @click="handleFilterChange(option.value)"
        >
          {{ option.label }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/vue-ui';

interface FilterOption {
  value: string;
  label: string;
}

interface ServicesFilterProps {
  filterOptions?: FilterOption[];
  onFilterChange?: (filter: string) => void;
}

const props = withDefaults(defineProps<ServicesFilterProps>(), {
  filterOptions: () => [{ value: 'all', label: 'Show All' }],
});

const activeFilter = ref<string>('all');

const handleFilterChange = (filter: string) => {
  activeFilter.value = filter;

  // Apply the same filtering logic as the original buttons
  const serviceCards = document.querySelectorAll('.service-card');
  const filterButtons = document.querySelectorAll('.filter-btn');

  // Update button active states (for desktop)
  filterButtons.forEach(btn => btn.classList.remove('active'));
  const targetButton = document.querySelector(`[data-filter="${filter}"]`);
  if (targetButton) {
    targetButton.classList.add('active');
  }

  serviceCards.forEach(card => {
    let show = false;

    if (filter === 'all') {
      show = true;
    } else {
      // Check if the card has the data attribute for this delivery mode
      show = card.getAttribute(`data-${filter}`) === 'true';
    }

    if (show) {
      card.classList.remove('hidden');
    } else {
      card.classList.add('hidden');
    }
  });

  // Check if any categories are empty and hide them
  document.querySelectorAll('.service-category').forEach(category => {
    const visibleCards = category.querySelectorAll('.service-card:not(.hidden)');
    if (visibleCards.length === 0) {
      (category as HTMLElement).style.display = 'none';
    } else {
      (category as HTMLElement).style.display = 'block';
    }
  });

  if (props.onFilterChange) {
    props.onFilterChange(filter);
  }
};

const handleButtonClick = (event: Event) => {
  const button = event.target as HTMLButtonElement;
  const filter = button.getAttribute('data-filter');
  if (filter) {
    activeFilter.value = filter;
  }
};

onMounted(() => {
  const filterButtons = document.querySelectorAll('.filter-btn');

  filterButtons.forEach(button => {
    button.addEventListener('click', handleButtonClick);
  });

  // Initialize the default filter (Show All) on component mount
  handleFilterChange('all');
});

onUnmounted(() => {
  const filterButtons = document.querySelectorAll('.filter-btn');
  filterButtons.forEach(button => {
    button.removeEventListener('click', handleButtonClick);
  });
});
</script>
