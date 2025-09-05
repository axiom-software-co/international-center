<template>
  <div :class="containerClasses">
    <!-- Skeleton loading for different content types -->
    <div v-if="variant === 'card'" class="space-y-6">
      <div v-for="n in items" :key="n" class="animate-pulse">
        <div class="border border-gray-200 dark:border-gray-700 rounded-lg p-6">
          <div class="flex items-start justify-between mb-4">
            <div class="h-6 bg-gray-300 dark:bg-gray-600 rounded w-3/4"></div>
            <div class="h-5 bg-gray-300 dark:bg-gray-600 rounded w-16"></div>
          </div>
          <div class="space-y-2 mb-4">
            <div class="h-4 bg-gray-300 dark:bg-gray-600 rounded w-full"></div>
            <div class="h-4 bg-gray-300 dark:bg-gray-600 rounded w-5/6"></div>
          </div>
          <div class="space-y-2">
            <div class="h-3 bg-gray-300 dark:bg-gray-600 rounded w-2/3"></div>
            <div class="h-3 bg-gray-300 dark:bg-gray-600 rounded w-3/4"></div>
          </div>
        </div>
      </div>
    </div>

    <!-- List item loading -->
    <div v-else-if="variant === 'list'" class="space-y-4">
      <div v-for="n in items" :key="n" class="animate-pulse">
        <div class="flex items-center space-x-4 p-4">
          <div class="h-10 w-10 bg-gray-300 dark:bg-gray-600 rounded-full"></div>
          <div class="flex-1 space-y-2">
            <div class="h-4 bg-gray-300 dark:bg-gray-600 rounded w-3/4"></div>
            <div class="h-3 bg-gray-300 dark:bg-gray-600 rounded w-1/2"></div>
          </div>
          <div class="h-4 bg-gray-300 dark:bg-gray-600 rounded w-16"></div>
        </div>
      </div>
    </div>

    <!-- Table loading -->
    <div v-else-if="variant === 'table'" class="animate-pulse">
      <div class="bg-white dark:bg-gray-800 shadow overflow-hidden rounded-lg">
        <!-- Table header -->
        <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
          <div class="flex space-x-4">
            <div class="h-4 bg-gray-300 dark:bg-gray-600 rounded w-24"></div>
            <div class="h-4 bg-gray-300 dark:bg-gray-600 rounded w-32"></div>
            <div class="h-4 bg-gray-300 dark:bg-gray-600 rounded w-20"></div>
            <div class="h-4 bg-gray-300 dark:bg-gray-600 rounded w-16"></div>
          </div>
        </div>
        <!-- Table rows -->
        <div v-for="n in items" :key="n" class="px-6 py-4 border-b border-gray-200 dark:border-gray-700 last:border-b-0">
          <div class="flex space-x-4">
            <div class="h-4 bg-gray-300 dark:bg-gray-600 rounded w-24"></div>
            <div class="h-4 bg-gray-300 dark:bg-gray-600 rounded w-32"></div>
            <div class="h-4 bg-gray-300 dark:bg-gray-600 rounded w-20"></div>
            <div class="h-4 bg-gray-300 dark:bg-gray-600 rounded w-16"></div>
          </div>
        </div>
      </div>
    </div>

    <!-- Article/content loading -->
    <div v-else-if="variant === 'article'" class="animate-pulse space-y-6">
      <div class="space-y-2">
        <div class="h-8 bg-gray-300 dark:bg-gray-600 rounded w-3/4"></div>
        <div class="h-4 bg-gray-300 dark:bg-gray-600 rounded w-1/2"></div>
      </div>
      <div class="h-48 bg-gray-300 dark:bg-gray-600 rounded"></div>
      <div class="space-y-4">
        <div class="space-y-2">
          <div class="h-4 bg-gray-300 dark:bg-gray-600 rounded w-full"></div>
          <div class="h-4 bg-gray-300 dark:bg-gray-600 rounded w-full"></div>
          <div class="h-4 bg-gray-300 dark:bg-gray-600 rounded w-3/4"></div>
        </div>
        <div class="space-y-2">
          <div class="h-4 bg-gray-300 dark:bg-gray-600 rounded w-full"></div>
          <div class="h-4 bg-gray-300 dark:bg-gray-600 rounded w-5/6"></div>
        </div>
      </div>
    </div>

    <!-- Form loading -->
    <div v-else-if="variant === 'form'" class="animate-pulse space-y-6">
      <div v-for="n in Math.min(items, 6)" :key="n" class="space-y-2">
        <div class="h-4 bg-gray-300 dark:bg-gray-600 rounded w-24"></div>
        <div class="h-10 bg-gray-300 dark:bg-gray-600 rounded w-full"></div>
      </div>
      <div class="flex space-x-4">
        <div class="h-10 bg-gray-300 dark:bg-gray-600 rounded w-24"></div>
        <div class="h-10 bg-gray-300 dark:bg-gray-600 rounded w-20"></div>
      </div>
    </div>

    <!-- Spinner loading (default) -->
    <div v-else class="flex items-center justify-center">
      <div class="flex flex-col items-center space-y-3">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        <p v-if="message" class="text-sm text-gray-600 dark:text-gray-400">{{ message }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';

type LoadingVariant = 'card' | 'list' | 'table' | 'article' | 'form' | 'spinner';

interface LoadingStateProps {
  variant?: LoadingVariant;
  items?: number;
  message?: string;
  fullHeight?: boolean;
  centered?: boolean;
}

const props = withDefaults(defineProps<LoadingStateProps>(), {
  variant: 'spinner',
  items: 3,
  message: '',
  fullHeight: false,
  centered: true,
});

const containerClasses = computed(() => [
  'loading-state',
  {
    'min-h-screen': props.fullHeight,
    'min-h-32': !props.fullHeight && props.variant === 'spinner',
    'flex items-center justify-center': props.centered && props.variant === 'spinner',
    'py-4': props.variant !== 'spinner',
  }
]);
</script>

<style scoped>
.loading-state {
  width: 100%;
}

/* Improve animation performance */
.animate-pulse {
  animation-duration: 2s;
}

@media (prefers-reduced-motion: reduce) {
  .animate-pulse,
  .animate-spin {
    animation: none;
  }
}

/* Custom pulse animation for better performance */
@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

.animate-pulse {
  animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}
</style>